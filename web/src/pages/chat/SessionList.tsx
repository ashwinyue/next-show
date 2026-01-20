import { useEffect } from 'react';
import { useChatStore } from '@/stores/useChatStore';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Plus, MessageSquare, Trash2 } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { format } from 'date-fns';

export function SessionList() {
  const { sessions, currentSessionId, fetchSessions, createSession, deleteSession, selectSession } = useChatStore();

  useEffect(() => {
    fetchSessions();
  }, []);

  const handleCreateSession = async () => {
    await createSession("builtin-deep"); // Default agent
  };

  const handleDeleteSession = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    if (confirm('Are you sure you want to delete this session?')) {
      await deleteSession(id);
    }
  };

  return (
    <div className="flex h-full w-64 flex-col border-r bg-muted/20">
      <div className="p-4">
        <Button onClick={handleCreateSession} className="w-full justify-start gap-2" variant="outline">
          <Plus className="h-4 w-4" />
          New Chat
        </Button>
      </div>
      <ScrollArea className="flex-1 px-2">
        <div className="space-y-1 p-2">
          {sessions.map((session) => (
            <div
              key={session.id}
              onClick={() => selectSession(session.id)}
              className={cn(
                "group flex w-full items-center justify-between rounded-md border border-transparent px-3 py-2 text-sm transition-colors hover:bg-accent hover:text-accent-foreground cursor-pointer",
                currentSessionId === session.id ? "bg-accent text-accent-foreground font-medium" : "text-muted-foreground"
              )}
            >
              <div className="flex items-center gap-2 overflow-hidden">
                <MessageSquare className="h-4 w-4 shrink-0" />
                <div className="flex flex-col overflow-hidden text-left">
                   <span className="truncate">{session.title || 'Untitled Session'}</span>
                   <span className="text-[10px] opacity-60 truncate">{format(new Date(session.updated_at), 'MM/dd HH:mm')}</span>
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                onClick={(e) => handleDeleteSession(e, session.id)}
              >
                <Trash2 className="h-3 w-3 text-destructive" />
              </Button>
            </div>
          ))}
          {sessions.length === 0 && (
            <div className="text-center text-xs text-muted-foreground py-8">
              No history
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
