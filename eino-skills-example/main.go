package main

import (
	"context"
	"eino-skills-example/disk_filesystem"
	"fmt"
	"github.com/cloudwego/eino-examples/adk/common/prints"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/filesystem"
	eino_skill "github.com/cloudwego/eino/adk/middlewares/skill"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	localBackendConfig := eino_skill.LocalBackendConfig{
		BaseDir: os.Getenv("BASE_DIR"),
	}
	localBackend, err := eino_skill.NewLocalBackend(&localBackendConfig)
	if err != nil {
		log.Fatal(err)
	}
	name := "load_skill"
	skillMiddleware, err := eino_skill.New(ctx, &eino_skill.Config{Backend: localBackend, SkillToolName: &name, UseChinese: true})
	if err != nil {
		log.Fatal(err)
	}
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   os.Getenv("OPENAI_MODEL"),
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		log.Fatal(err)
	}

	toolsSystemPrompt := fmt.Sprintf(`
		# Filesystem Tools 'ls', 'read_file', 'write_file', 'edit_file', 'glob', 'grep'
		You have access to a filesystem which you can interact with using these tools.
		All file paths must start with a '%s'.
		- ls: list files in a directory (requires absolute path)
		- read_file: read a file from the filesystem
		- write_file: write to a file in the filesystem
		- edit_file: edit a file in the filesystem
		- glob: find files matching a pattern (e.g., "**/*.py")
		- grep: search for text within files
		`, os.Getenv("BASE_DIR"))
	filesystemConfig := &filesystem.Config{
		Backend:                          disk_filesystem.NewInDiskBackend(),
		CustomSystemPrompt:               &toolsSystemPrompt,
		WithoutLargeToolResultOffloading: true,
	}
	filesystemMiddleware, err := filesystem.NewMiddleware(ctx, filesystemConfig)
	if err != nil {
		log.Fatal(err)
	}
	chatModelAgentConfig := &adk.ChatModelAgentConfig{
		Name:          "skill-agent",
		Description:   "skill-agent",
		Model:         chatModel,
		Middlewares:   []adk.AgentMiddleware{skillMiddleware, filesystemMiddleware},
		MaxIterations: 100,
	}
	a, err := adk.NewChatModelAgent(ctx, chatModelAgentConfig)
	if err != nil {
		log.Fatal(err)
	}
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: true,
		Agent:           a,
	})

	//q := "我的产品是一个人工智能眼镜，目标为30-40岁的白领，眼镜主要用于嵌入AI功能，比如翻译，物体识别等等"
	q := `
	老陈：小李，下周二你跟我去趟上海，咱们得把那个大客户签下来。
	小李：没问题陈总，那我今天先把出差申请给报了。
	老陈：行，酒店你看着订，要方便出行的，外滩那边有个酒店不错，大概 1200 一晚。
	小李：1200 稍微有点贵，但我看地段确实好，那我就按这个金额报了？
	老陈：嗯，另外晚上咱们得请客户吃顿饭，规格得高一点。
	小李：明白。我预计 3000 块左右的包间，咱们一共 6 个人，这标准行吗？
	老陈：行，人均 500 在上海这地方也算正常，为了签单这钱该花。
	小李：好，那我申请单里的住宿费填 1200，餐饮费填 3000，我待会直接提交系统。
	老陈：可以，你动作快点，审批完了咱们好赶紧订票。
	总结上述会议纪要`

	iter := runner.Run(ctx, []adk.Message{
		{Role: "user", Content: q}})
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatal(event.Err)
		}
		prints.Event(event)
	}
}
