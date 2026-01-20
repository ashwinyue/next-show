import { useState, useRef, useEffect } from 'react';
import { useChatStore, type Message, type Reference } from '@/stores/useChatStore';
import { useAgentStore } from '@/stores/useAgentStore';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Send, User, Bot, Loader2, ChevronDown, ChevronRight, Terminal, Copy, Check, BookOpen, Paperclip } from 'lucide-react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { SessionList } from './SessionList';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { useToast } from '@/hooks/use-toast';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import 'highlight.js/styles/github-dark.css'; 

export function ChatPage() {
  const { messages, isLoading, sendMessage, currentSessionId } = useChatStore();
  const { agents, fetchAgents } = useAgentStore();
  const [input, setInput] = useState('');
  const [selectedAgentId, setSelectedAgentId] = useState<string>('builtin-deep');
  const scrollRef = useRef<HTMLDivElement>(null);
  const { toast } = useToast();

  useEffect(() => {
    fetchAgents();
  }, []);

  const handleSend = async () => {
    if (!input.trim() || isLoading) return;
    const content = input;
    setInput('');
    await sendMessage(content, selectedAgentId);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleFileUpload = () => {
    toast({
        title: "File Upload",
        description: "Please upload documents in the Knowledge Base to chat with them.",
    });
  };

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <div className="flex h-full bg-background overflow-hidden">
      {/* Session Sidebar */}
      <SessionList />

      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col h-full relative">
        {/* Header */}
        <div className="border-b px-6 py-4 flex items-center justify-between bg-card/50 backdrop-blur-sm sticky top-0 z-10">
            <div>
            <h1 className="text-xl font-bold font-mono tracking-tight">Chat Session</h1>
            <p className="text-xs text-muted-foreground font-mono">
                {currentSessionId ? `ID: ${currentSessionId.slice(0,8)}...` : 'New Session'}
            </p>
            </div>
            <div className="flex items-center gap-2">
                <Select value={selectedAgentId} onValueChange={setSelectedAgentId}>
                    <SelectTrigger className="w-[180px] h-8 text-xs font-mono">
                        <SelectValue placeholder="Select Agent" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="builtin-deep">Builtin DeepSeek</SelectItem>
                        {agents.map((agent) => (
                            <SelectItem key={agent.id} value={agent.id}>
                                {agent.display_name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>
        </div>

        {/* Messages Area */}
        <div className="flex-1 overflow-hidden relative">
            <div className="absolute inset-0 overflow-y-auto p-4 space-y-6" ref={scrollRef}>
                {messages.length === 0 ? (
                    <div className="h-full flex flex-col items-center justify-center text-muted-foreground opacity-50">
                        <Bot className="h-16 w-16 mb-4" />
                        <p className="font-mono">Ready to start conversation.</p>
                    </div>
                ) : (
                    messages.map((msg) => (
                        <ChatBubble key={msg.id} message={msg} />
                    ))
                )}
                {isLoading && messages[messages.length - 1]?.role === 'user' && (
                    <div className="flex items-center gap-2 text-muted-foreground p-4 animate-pulse">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        <span className="text-xs font-mono">Thinking...</span>
                    </div>
                )}
            </div>
        </div>

        {/* Input Area */}
        <div className="p-4 border-t bg-card">
            <div className="relative max-w-4xl mx-auto">
                <Textarea
                value={input}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Type your message..."
                className="min-h-[80px] w-full resize-none pr-24 bg-background/50 font-sans shadow-sm focus-visible:ring-primary pt-3"
                />
                <div className="absolute bottom-2 right-2 flex gap-1">
                    <Button
                        size="icon"
                        variant="ghost"
                        className="h-8 w-8 text-muted-foreground hover:text-foreground"
                        onClick={handleFileUpload}
                        title="Upload File"
                    >
                        <Paperclip className="h-4 w-4" />
                    </Button>
                    <Button 
                        size="icon" 
                        className="h-8 w-8 transition-transform active:scale-95"
                        onClick={handleSend}
                        disabled={isLoading || !input.trim()}
                    >
                        <Send className="h-4 w-4" />
                    </Button>
                </div>
            </div>
            <div className="text-center mt-2">
                <p className="text-[10px] text-muted-foreground font-mono">AI can make mistakes. Check important info.</p>
            </div>
        </div>
      </div>
    </div>
  );
}

function ChatBubble({ message }: { message: Message }) {
  const isUser = message.role === 'user';
  const [isThinkingOpen, setIsThinkingOpen] = useState(true);
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(message.content);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className={cn("flex gap-4 max-w-4xl mx-auto group", isUser ? "flex-row-reverse" : "flex-row")}>
      <Avatar className={cn("h-8 w-8 border shrink-0", isUser ? "bg-primary text-primary-foreground" : "bg-muted")}>
        <AvatarImage src={isUser ? undefined : "/bot-avatar.png"} />
        <AvatarFallback>{isUser ? <User className="h-4 w-4" /> : <Bot className="h-4 w-4" />}</AvatarFallback>
      </Avatar>

      <div className={cn("flex-1 space-y-2 min-w-0", isUser && "text-right")}>
        <div className={cn(
            "inline-block rounded-lg p-3 text-sm shadow-sm border text-left max-w-full overflow-hidden",
            isUser ? "bg-primary text-primary-foreground border-primary" : "bg-card border-border"
        )}>
           {/* Thinking Process */}
           {!isUser && message.thinking && (
               <div className="mb-3">
                   <button 
                     onClick={() => setIsThinkingOpen(!isThinkingOpen)}
                     className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors font-mono w-full text-left bg-muted/30 p-1.5 rounded-md border border-transparent hover:border-border"
                   >
                       {isThinkingOpen ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                       Thinking Process
                   </button>
                   {isThinkingOpen && (
                       <div className="mt-1 pl-3 border-l-2 border-primary/20 text-xs text-muted-foreground font-mono whitespace-pre-wrap animate-in fade-in slide-in-from-top-1 bg-muted/10 p-2 rounded-r-md">
                           {message.thinking}
                       </div>
                   )}
               </div>
           )}

           {/* Tool Calls */}
           {!isUser && message.toolCalls && message.toolCalls.length > 0 && (
               <div className="mb-3 space-y-2">
                   {message.toolCalls.map((tool, idx) => (
                       <div key={idx} className="bg-black/90 text-green-400 p-2.5 rounded-md text-xs font-mono border border-green-900/50 flex flex-col gap-1 shadow-inner">
                           <div className="flex items-center gap-2 border-b border-green-900/30 pb-1 mb-1">
                               <Terminal className="h-3 w-3" />
                               <span className="font-bold">{tool.name}</span>
                               <span className={cn("ml-auto text-[10px] px-1.5 py-0.5 rounded", 
                                   tool.status === 'success' ? 'bg-green-900/50 text-green-300' : 'bg-yellow-900/50 text-yellow-300'
                               )}>
                                   {tool.status}
                               </span>
                           </div>
                           <div className="opacity-90 break-all whitespace-pre-wrap">
                               {JSON.stringify(tool.arguments, null, 2)}
                           </div>
                           {tool.result && (
                               <div className="mt-1 pt-1 border-t border-green-900/30 text-blue-300">
                                   → {JSON.stringify(tool.result).slice(0, 100)}...
                               </div>
                           )}
                       </div>
                   ))}
               </div>
           )}

           {/* References */}
           {!isUser && message.references && message.references.length > 0 && (
               <div className="mb-3">
                   <div className="text-xs font-semibold text-muted-foreground mb-1 flex items-center gap-1">
                       <BookOpen className="h-3 w-3" />
                       References
                   </div>
                   <div className="flex flex-wrap gap-2">
                       {message.references.map((ref: Reference) => (
                           <Popover key={ref.id}>
                               <PopoverTrigger asChild>
                                   <Button variant="outline" size="sm" className="h-6 text-[10px] px-2 bg-background/50 border-dashed">
                                       {ref.title || 'Document'} 
                                       {ref.score && <span className="ml-1 opacity-50">({ref.score.toFixed(2)})</span>}
                                   </Button>
                               </PopoverTrigger>
                               <PopoverContent className="w-80 text-xs p-3">
                                   <div className="font-semibold mb-1">{ref.title}</div>
                                   <div className="text-muted-foreground max-h-40 overflow-y-auto whitespace-pre-wrap">
                                       {ref.content_preview}
                                   </div>
                               </PopoverContent>
                           </Popover>
                       ))}
                   </div>
               </div>
           )}

           {/* Main Content */}
           <div className={cn("prose prose-sm dark:prose-invert max-w-none break-words leading-relaxed", isUser && "text-primary-foreground")}>
               <ReactMarkdown 
                 remarkPlugins={[remarkGfm]}
                 rehypePlugins={[rehypeHighlight]}
                 components={{
                     pre: ({node, ref, ...props}) => <pre className="overflow-auto my-2 rounded-md bg-muted/50 p-0" {...props} />,
                     code: ({node, className, children, ...props}) => {
                         return !String(className).includes('language-') ? (
                             <code className={cn("bg-muted/50 px-1 py-0.5 rounded font-mono text-xs", className)} {...props}>
                                 {children}
                             </code>
                         ) : (
                             <code className={cn("block p-3 font-mono text-xs", className)} {...props}>
                                 {children}
                             </code>
                         )
                     }
                 }}
               >
                   {message.content}
               </ReactMarkdown>
           </div>
        </div>
        
        {/* Message Footer */}
        <div className="flex items-center justify-between px-1">
             <div className="text-[10px] text-muted-foreground font-mono opacity-50">
                {new Date(message.createdAt).toLocaleTimeString()}
                {message.status === 'error' && <span className="text-destructive ml-2">Error: {message.error}</span>}
             </div>
             {!isUser && (
                 <Button variant="ghost" size="icon" className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity" onClick={handleCopy}>
                     {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
                 </Button>
             )}
        </div>
      </div>
    </div>
  );
}
