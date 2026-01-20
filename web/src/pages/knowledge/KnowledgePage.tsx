import { useEffect, useState } from 'react';
import * as React from 'react';
import { useKnowledgeStore } from '@/stores/useKnowledgeStore';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Database, Plus, Trash2, FileText, Upload, ChevronLeft, Layers } from 'lucide-react';
import { format } from 'date-fns';
import { Badge } from '@/components/ui/badge';

export function KnowledgePage() {
  const { 
    knowledgeBases, fetchKnowledgeBases, createKnowledgeBase, deleteKnowledgeBase, selectKnowledgeBase, selectedKb,
    documents, uploadDocument, deleteDocument,
    chunks, fetchChunks
  } = useKnowledgeStore();

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newKbName, setNewKbName] = useState('');
  const [newKbDesc, setNewKbDesc] = useState('');
  
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [selectedDocId, setSelectedDocId] = useState<string | null>(null);

  useEffect(() => {
    fetchKnowledgeBases();
  }, []);

  const handleCreateKb = async () => {
    if (!newKbName) return;
    await createKnowledgeBase({
      name: newKbName,
      description: newKbDesc,
      embedding_model: 'text-embedding-3-small', // Hardcoded
    });
    setIsCreateOpen(false);
    setNewKbName('');
    setNewKbDesc('');
  };

  const handleUpload = async () => {
    if (!selectedKb || !uploadFile) return;
    await uploadDocument(selectedKb.id, uploadFile);
    setUploadFile(null);
  };

  const handleViewChunks = async (docId: string) => {
    if (selectedDocId === docId) {
        setSelectedDocId(null);
    } else {
        setSelectedDocId(docId);
        await fetchChunks(docId);
    }
  };

  if (selectedKb) {
    return (
      <div className="p-8 h-full flex flex-col overflow-hidden">
        <div className="flex items-center gap-4 mb-6">
          <Button variant="ghost" size="icon" onClick={() => selectKnowledgeBase(null)}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold font-mono flex items-center gap-2">
              <Database className="h-6 w-6 text-primary" />
              {selectedKb.name}
            </h1>
            <p className="text-muted-foreground">{selectedKb.description}</p>
          </div>
        </div>

        <Tabs defaultValue="documents" className="flex-1 flex flex-col overflow-hidden">
          <TabsList>
            <TabsTrigger value="documents">Documents</TabsTrigger>
            <TabsTrigger value="search">Search Test</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
          
          <TabsContent value="documents" className="flex-1 overflow-hidden flex flex-col mt-4">
             <div className="flex justify-between items-center mb-4">
                 <h2 className="text-lg font-semibold">Managed Documents</h2>
                 <div className="flex items-center gap-2">
                     <Input 
                        type="file" 
                        className="w-64" 
                        onChange={(e) => setUploadFile(e.target.files?.[0] || null)}
                     />
                     <Button disabled={!uploadFile} onClick={handleUpload}>
                         <Upload className="h-4 w-4 mr-2" />
                         Upload
                     </Button>
                 </div>
             </div>
             
             <div className="flex-1 overflow-auto border rounded-md">
                 <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Title</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Chunks</TableHead>
                            <TableHead>Created At</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {documents.map((doc) => (
                            <React.Fragment key={doc.id}>
                                <TableRow className="cursor-pointer hover:bg-muted/50" onClick={() => handleViewChunks(doc.id)}>
                                    <TableCell className="font-medium flex items-center gap-2">
                                        <FileText className="h-4 w-4 text-muted-foreground" />
                                        {doc.title}
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant={doc.status === 'completed' ? 'default' : 'secondary'}>
                                            {doc.status}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>{doc.chunk_count}</TableCell>
                                    <TableCell className="text-xs text-muted-foreground">
                                        {format(new Date(doc.created_at), 'MMM dd HH:mm')}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <Button 
                                            variant="ghost" 
                                            size="icon" 
                                            onClick={(e) => { e.stopPropagation(); deleteDocument(doc.id); }}
                                            className="text-destructive hover:text-destructive"
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                                {selectedDocId === doc.id && (
                                    <TableRow className="bg-muted/30 hover:bg-muted/30">
                                        <TableCell colSpan={5} className="p-4">
                                            <div className="rounded-md border bg-background p-4">
                                                <h3 className="font-semibold mb-2 flex items-center gap-2">
                                                    <Layers className="h-4 w-4" />
                                                    Document Chunks
                                                </h3>
                                                <div className="space-y-2 max-h-64 overflow-y-auto">
                                                    {chunks.map((chunk) => (
                                                        <div key={chunk.id} className="text-xs font-mono p-2 border rounded bg-muted/50">
                                                            <div className="mb-1 text-muted-foreground">Chunk #{chunk.index} ({chunk.token_count} tokens)</div>
                                                            <div className="break-all">{chunk.content}</div>
                                                        </div>
                                                    ))}
                                                    {chunks.length === 0 && <div className="text-muted-foreground text-sm">No chunks found.</div>}
                                                </div>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                )}
                            </React.Fragment>
                        ))}
                        {documents.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                                    No documents uploaded yet.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                 </Table>
             </div>
          </TabsContent>
          
          <TabsContent value="search">
              <div className="flex items-center justify-center h-64 text-muted-foreground">
                  Search test interface coming soon.
              </div>
          </TabsContent>
        </Tabs>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6 h-full overflow-y-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold font-mono tracking-tight">Knowledge Base</h1>
          <p className="text-muted-foreground mt-2">Manage your document collections for RAG.</p>
        </div>

        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              New Knowledge Base
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Knowledge Base</DialogTitle>
              <DialogDescription>
                Create a new collection to store and index your documents.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="name" className="text-right">Name</Label>
                <Input
                  id="name"
                  value={newKbName}
                  onChange={(e) => setNewKbName(e.target.value)}
                  className="col-span-3"
                  placeholder="Company Wiki"
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="desc" className="text-right">Description</Label>
                <Textarea
                  id="desc"
                  value={newKbDesc}
                  onChange={(e) => setNewKbDesc(e.target.value)}
                  className="col-span-3"
                />
              </div>
            </div>
            <DialogFooter>
              <Button onClick={handleCreateKb}>Create</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {knowledgeBases.map((kb) => (
          <Card 
            key={kb.id} 
            className="group cursor-pointer hover:shadow-md hover:border-primary/50 transition-all relative overflow-hidden"
            onClick={() => selectKnowledgeBase(kb)}
          >
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg flex items-center gap-2">
                  <Database className="h-5 w-5 text-primary" />
                  {kb.name}
                </CardTitle>
              </div>
              <CardDescription className="line-clamp-2 mt-2">
                {kb.description || 'No description provided.'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-xs text-muted-foreground font-mono">
                Model: {kb.embedding_model}
              </div>
              <div className="text-xs text-muted-foreground font-mono mt-1">
                Created: {format(new Date(kb.created_at), 'yyyy-MM-dd')}
              </div>
            </CardContent>
            <div className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity">
                <Button 
                    variant="ghost" 
                    size="icon" 
                    className="h-8 w-8 text-destructive hover:text-destructive"
                    onClick={(e) => { e.stopPropagation(); deleteKnowledgeBase(kb.id); }}
                >
                    <Trash2 className="h-4 w-4" />
                </Button>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
