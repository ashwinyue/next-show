// Package tools 提供内置工具和中间件.
package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	_ "github.com/marcboeker/go-duckdb"
)

// DataAnalysisManager 管理 DuckDB 数据分析会话.
type DataAnalysisManager struct {
	db            *sql.DB
	mu            sync.Mutex
	loadedTables  map[string]string   // documentID -> tableName
	sessionTables map[string][]string // sessionID -> []tableName (用于清理)
}

// NewDataAnalysisManager 创建数据分析管理器.
func NewDataAnalysisManager() (*DataAnalysisManager, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}

	return &DataAnalysisManager{
		db:            db,
		loadedTables:  make(map[string]string),
		sessionTables: make(map[string][]string),
	}, nil
}

// Close 关闭数据库连接.
func (m *DataAnalysisManager) Close() error {
	return m.db.Close()
}

// LoadCSVFile 加载 CSV 文件到 DuckDB.
func (m *DataAnalysisManager) LoadCSVFile(ctx context.Context, sessionID, documentID, filePath string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已加载
	if tableName, ok := m.loadedTables[documentID]; ok {
		return tableName, nil
	}

	// 生成表名（使用 document ID 的前 8 位）
	tableName := fmt.Sprintf("doc_%s", strings.ReplaceAll(documentID, "-", "")[:8])

	// 创建表
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" AS SELECT * FROM read_csv_auto('%s', header=true)`, tableName, filePath)
	if _, err := m.db.ExecContext(ctx, query); err != nil {
		return "", fmt.Errorf("load csv: %w", err)
	}

	m.loadedTables[documentID] = tableName
	m.sessionTables[sessionID] = append(m.sessionTables[sessionID], tableName)

	return tableName, nil
}

// LoadXLSXFile 加载 XLSX 文件到 DuckDB.
func (m *DataAnalysisManager) LoadXLSXFile(ctx context.Context, sessionID, documentID, filePath string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已加载
	if tableName, ok := m.loadedTables[documentID]; ok {
		return tableName, nil
	}

	// 生成表名
	tableName := fmt.Sprintf("doc_%s", strings.ReplaceAll(documentID, "-", "")[:8])

	// DuckDB 需要安装 spatial 扩展来读取 xlsx
	// 先尝试安装扩展
	m.db.ExecContext(ctx, "INSTALL spatial; LOAD spatial;")

	// 创建表
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" AS SELECT * FROM st_read('%s')`, tableName, filePath)
	if _, err := m.db.ExecContext(ctx, query); err != nil {
		return "", fmt.Errorf("load xlsx: %w", err)
	}

	m.loadedTables[documentID] = tableName
	m.sessionTables[sessionID] = append(m.sessionTables[sessionID], tableName)

	return tableName, nil
}

// GetTableSchema 获取表结构信息.
func (m *DataAnalysisManager) GetTableSchema(ctx context.Context, tableName string) (*TableSchema, error) {
	// 获取列信息
	query := fmt.Sprintf(`DESCRIBE "%s"`, tableName)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("describe table: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var colName, colType, null, key, defaultVal, extra sql.NullString
		if err := rows.Scan(&colName, &colType, &null, &key, &defaultVal, &extra); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}
		columns = append(columns, ColumnInfo{
			Name: colName.String,
			Type: colType.String,
		})
	}

	// 获取行数
	var rowCount int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tableName)
	if err := m.db.QueryRowContext(ctx, countQuery).Scan(&rowCount); err != nil {
		return nil, fmt.Errorf("count rows: %w", err)
	}

	return &TableSchema{
		TableName: tableName,
		Columns:   columns,
		RowCount:  rowCount,
	}, nil
}

// ExecuteQuery 执行只读 SQL 查询.
func (m *DataAnalysisManager) ExecuteQuery(ctx context.Context, sqlQuery string, maxRows int) (*QueryResult, error) {
	// 安全检查：只允许 SELECT 查询
	normalized := strings.TrimSpace(strings.ToLower(sqlQuery))
	if !strings.HasPrefix(normalized, "select") &&
		!strings.HasPrefix(normalized, "show") &&
		!strings.HasPrefix(normalized, "describe") &&
		!strings.HasPrefix(normalized, "explain") {
		return nil, fmt.Errorf("only SELECT queries are allowed")
	}

	// 添加 LIMIT 如果没有
	if !strings.Contains(normalized, "limit") && maxRows > 0 {
		sqlQuery = fmt.Sprintf("%s LIMIT %d", sqlQuery, maxRows)
	}

	rows, err := m.db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	// 获取列名
	colNames, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	// 读取数据
	var data []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(colNames))
		valuePtrs := make([]interface{}, len(colNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range colNames {
			row[col] = values[i]
		}
		data = append(data, row)
	}

	return &QueryResult{
		Columns:  colNames,
		Data:     data,
		RowCount: len(data),
	}, nil
}

// CleanupSession 清理会话相关的表.
func (m *DataAnalysisManager) CleanupSession(ctx context.Context, sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tables, ok := m.sessionTables[sessionID]
	if !ok {
		return
	}

	for _, tableName := range tables {
		m.db.ExecContext(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, tableName))
		// 从 loadedTables 中移除
		for docID, tbl := range m.loadedTables {
			if tbl == tableName {
				delete(m.loadedTables, docID)
				break
			}
		}
	}

	delete(m.sessionTables, sessionID)
}

// TableSchema 表结构信息.
type TableSchema struct {
	TableName string       `json:"table_name"`
	Columns   []ColumnInfo `json:"columns"`
	RowCount  int64        `json:"row_count"`
}

// ColumnInfo 列信息.
type ColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// QueryResult 查询结果.
type QueryResult struct {
	Columns  []string                 `json:"columns"`
	Data     []map[string]interface{} `json:"data"`
	RowCount int                      `json:"row_count"`
}

// DataSchemaTool 数据表结构查询工具.
type DataSchemaTool struct {
	manager     *DataAnalysisManager
	getFilePath func(ctx context.Context, documentID string) (string, string, error) // returns (filePath, fileType, error)
	sessionID   string
}

// DataSchemaInput 数据表结构查询输入.
type DataSchemaInput struct {
	DocumentID string `json:"document_id" jsonschema:"description=要查询结构的文档 ID"`
}

// NewDataSchemaTool 创建数据表结构查询工具.
func NewDataSchemaTool(manager *DataAnalysisManager, sessionID string, getFilePath func(ctx context.Context, documentID string) (string, string, error)) tool.InvokableTool {
	t := &DataSchemaTool{
		manager:     manager,
		getFilePath: getFilePath,
		sessionID:   sessionID,
	}
	return t
}

func (t *DataSchemaTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "data_schema",
		Desc: "获取 CSV/Excel 文件的表结构信息（表名、列名、列类型、行数）。在执行任何 SQL 查询之前，必须先调用此工具了解数据结构。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"document_id": {
				Type: schema.String,
				Desc: "要查询结构的文档 ID",
			},
		}),
	}, nil
}

func (t *DataSchemaTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input DataSchemaInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse input: %w", err)
	}

	// 获取文件路径
	filePath, fileType, err := t.getFilePath(ctx, input.DocumentID)
	if err != nil {
		return "", fmt.Errorf("get file path: %w", err)
	}

	// 加载文件到 DuckDB
	var tableName string
	switch fileType {
	case "csv":
		tableName, err = t.manager.LoadCSVFile(ctx, t.sessionID, input.DocumentID, filePath)
	case "xlsx", "xls":
		tableName, err = t.manager.LoadXLSXFile(ctx, t.sessionID, input.DocumentID, filePath)
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}
	if err != nil {
		return "", err
	}

	// 获取表结构
	schema, err := t.manager.GetTableSchema(ctx, tableName)
	if err != nil {
		return "", err
	}

	// 格式化输出
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 表结构信息\n\n"))
	sb.WriteString(fmt.Sprintf("**表名**: `%s`\n", schema.TableName))
	sb.WriteString(fmt.Sprintf("**行数**: %d\n\n", schema.RowCount))
	sb.WriteString("**列信息**:\n\n")
	sb.WriteString("| 列名 | 类型 |\n")
	sb.WriteString("|------|------|\n")
	for _, col := range schema.Columns {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", col.Name, col.Type))
	}

	return sb.String(), nil
}

// DataAnalysisTool 数据分析 SQL 查询工具.
type DataAnalysisTool struct {
	manager *DataAnalysisManager
	maxRows int
}

// DataAnalysisInput 数据分析查询输入.
type DataAnalysisInput struct {
	SQL string `json:"sql" jsonschema:"description=要执行的 SQL 查询语句（只支持 SELECT）"`
}

// NewDataAnalysisTool 创建数据分析工具.
func NewDataAnalysisTool(manager *DataAnalysisManager, maxRows int) tool.InvokableTool {
	if maxRows <= 0 {
		maxRows = 100
	}
	return &DataAnalysisTool{
		manager: manager,
		maxRows: maxRows,
	}
}

func (t *DataAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "data_analysis",
		Desc: "对已加载的数据表执行 SQL 查询。只支持 SELECT 查询，禁止 INSERT/UPDATE/DELETE/CREATE/DROP 等修改操作。使用前必须先调用 data_schema 获取表结构。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"sql": {
				Type: schema.String,
				Desc: "要执行的 SQL 查询语句",
			},
		}),
	}, nil
}

func (t *DataAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input DataAnalysisInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse input: %w", err)
	}

	// 执行查询
	result, err := t.manager.ExecuteQuery(ctx, input.SQL, t.maxRows)
	if err != nil {
		return "", err
	}

	// 格式化输出为 Markdown 表格
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 查询结果\n\n"))
	sb.WriteString(fmt.Sprintf("**返回行数**: %d\n\n", result.RowCount))

	if len(result.Data) > 0 {
		// 表头
		sb.WriteString("| " + strings.Join(result.Columns, " | ") + " |\n")
		sb.WriteString("|" + strings.Repeat("---|", len(result.Columns)) + "\n")

		// 数据行
		for _, row := range result.Data {
			var values []string
			for _, col := range result.Columns {
				val := row[col]
				if val == nil {
					values = append(values, "NULL")
				} else {
					values = append(values, fmt.Sprintf("%v", val))
				}
			}
			sb.WriteString("| " + strings.Join(values, " | ") + " |\n")
		}
	} else {
		sb.WriteString("*无数据*\n")
	}

	return sb.String(), nil
}
