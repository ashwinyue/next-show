import { useEffect, useState } from 'react';
import { useAgentStore } from '@/stores/useAgentStore';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Bot, Plus, Settings2, Trash2 } from 'lucide-react';

export function AgentPage() {
  const { agents, fetchAgents, createAgent, deleteAgent } = useAgentStore();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    display_name: '',
    description: '',
    model_name: 'gpt-4o',
    agent_type: 'chat_model',
    system_prompt: '',
  });

  useEffect(() => {
    fetchAgents();
  }, []);

  const handleCreate = async () => {
    try {
      await createAgent({
        ...formData,
        provider_id: 'openai', // Hardcoded for now
        agent_role: 'specialist',
      } as any);
      setIsCreateOpen(false);
      setFormData({
        name: '',
        display_name: '',
        description: '',
        model_name: 'gpt-4o',
        agent_type: 'chat_model',
        system_prompt: '',
      });
    } catch (e) {
      // Error handling
    }
  };

  const handleDelete = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    if (confirm('Are you sure you want to delete this agent?')) {
      await deleteAgent(id);
    }
  };

  return (
    <div className="p-8 space-y-6 h-full overflow-y-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold font-mono tracking-tight">Agents</h1>
          <p className="text-muted-foreground mt-2">Manage your AI agents and their capabilities.</p>
        </div>
        
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Create Agent
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>Create New Agent</DialogTitle>
              <DialogDescription>
                Configure your new AI agent.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="name" className="text-right">Name (ID)</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="col-span-3"
                  placeholder="my-agent"
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="displayName" className="text-right">Display Name</Label>
                <Input
                  id="displayName"
                  value={formData.display_name}
                  onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                  className="col-span-3"
                  placeholder="My Assistant"
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="model" className="text-right">Model</Label>
                <Select
                  value={formData.model_name}
                  onValueChange={(value) => setFormData({ ...formData, model_name: value })}
                >
                  <SelectTrigger className="col-span-3">
                    <SelectValue placeholder="Select model" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="gpt-4o">GPT-4o</SelectItem>
                    <SelectItem value="gpt-3.5-turbo">GPT-3.5 Turbo</SelectItem>
                    <SelectItem value="deepseek-chat">DeepSeek Chat</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="description" className="text-right">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="col-span-3"
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="prompt" className="text-right">System Prompt</Label>
                <Textarea
                  id="prompt"
                  value={formData.system_prompt}
                  onChange={(e) => setFormData({ ...formData, system_prompt: e.target.value })}
                  className="col-span-3 font-mono text-xs"
                  rows={5}
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="submit" onClick={handleCreate}>Create</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {agents.map((agent) => (
          <Card key={agent.id} className="group relative overflow-hidden transition-all hover:shadow-md hover:border-primary/50 cursor-pointer">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                    <Bot className="h-5 w-5" />
                  </div>
                  <CardTitle className="text-lg">{agent.display_name}</CardTitle>
                </div>
                <Badge variant={agent.agent_role === 'orchestrator' ? 'default' : 'secondary'}>
                  {agent.agent_role}
                </Badge>
              </div>
              <CardDescription className="line-clamp-2 mt-2">
                {agent.description || 'No description provided.'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-2 text-xs text-muted-foreground font-mono">
                <span className="bg-muted px-1.5 py-0.5 rounded">{agent.model_name}</span>
                <span className="bg-muted px-1.5 py-0.5 rounded">{agent.agent_type}</span>
              </div>
            </CardContent>
            <div className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity flex gap-2">
               <Button variant="ghost" size="icon" className="h-8 w-8">
                 <Settings2 className="h-4 w-4" />
               </Button>
               <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive" onClick={(e) => handleDelete(e, agent.id)}>
                 <Trash2 className="h-4 w-4" />
               </Button>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
