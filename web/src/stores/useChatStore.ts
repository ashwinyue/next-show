import { create } from 'zustand';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/useAuthStore';

export interface ToolCall {
  id: string;
  name: string;
  arguments: any;
  status: 'calling' | 'success' | 'failed';
  result?: any;
}

export interface Reference {
  id: string;
  title: string;
  url?: string;
  content_preview?: string;
  score?: number;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  thinking?: string;
  toolCalls?: ToolCall[];
  references?: Reference[];
  createdAt: number;
  status?: 'sending' | 'streaming' | 'completed' | 'error';
  error?: string;
}

export interface Session {
  id: string;
  title: string;
  agent_id: string;
  updated_at: string;
  created_at: string;
}

interface ChatState {
  sessions: Session[];
  currentSessionId: string | null;
  messages: Message[];
  isLoading: boolean;
  isStreaming: boolean;
  
  // Actions
  fetchSessions: () => Promise<void>;
  createSession: (agentId: string) => Promise<string>;
  deleteSession: (sessionId: string) => Promise<void>;
  selectSession: (sessionId: string) => Promise<void>;
  sendMessage: (content: string, agentId: string) => Promise<void>;
  stopStream: () => void;
  clearMessages: () => void;
}

export const useChatStore = create<ChatState>((set, get) => {
  let abortController: AbortController | null = null;

  return {
    sessions: [],
    currentSessionId: null,
    messages: [],
    isLoading: false,
    isStreaming: false,

    fetchSessions: async () => {
      try {
        const response = await api.get('/sessions');
        set({ sessions: response.data.sessions || [] });
      } catch (error) {
        console.error('Failed to fetch sessions:', error);
      }
    },

    createSession: async (agentId: string) => {
      try {
        const response = await api.post('/sessions', { agent_id: agentId });
        const newSession = response.data;
        set((state) => ({ 
          sessions: [newSession, ...state.sessions],
          currentSessionId: newSession.id,
          messages: []
        }));
        return newSession.id;
      } catch (error) {
        console.error('Failed to create session:', error);
        throw error;
      }
    },

    deleteSession: async (sessionId: string) => {
      try {
        await api.delete(`/sessions/${sessionId}`);
        set((state) => ({
          sessions: state.sessions.filter(s => s.id !== sessionId),
          currentSessionId: state.currentSessionId === sessionId ? null : state.currentSessionId,
          messages: state.currentSessionId === sessionId ? [] : state.messages
        }));
      } catch (error) {
        console.error('Failed to delete session:', error);
      }
    },

    selectSession: async (sessionId: string) => {
      set({ currentSessionId: sessionId, isLoading: true });
      try {
        const response = await api.get(`/sessions/${sessionId}/messages`);
        // Transform backend messages to frontend format
        const loadedMessages: Message[] = (response.data.messages || []).map((m: any) => ({
          id: m.id,
          role: m.role,
          content: m.content,
          thinking: m.metadata?.thinking,
          toolCalls: m.metadata?.tool_calls,
          references: m.metadata?.references,
          createdAt: new Date(m.created_at).getTime(),
          status: 'completed'
        }));
        set({ messages: loadedMessages });
      } catch (error) {
        console.error('Failed to load messages:', error);
        set({ messages: [] });
      } finally {
        set({ isLoading: false });
      }
    },

    clearMessages: () => set({ messages: [] }),

    stopStream: () => {
      if (abortController) {
        abortController.abort();
        abortController = null;
      }
      set({ isStreaming: false });
    },

    sendMessage: async (content: string, agentId: string) => {
      const { currentSessionId } = get();
      const token = useAuthStore.getState().token;

      // 1. Ensure Session Exists
      let sessionId = currentSessionId;
      if (!sessionId) {
        try {
          sessionId = await get().createSession(agentId);
        } catch (e) {
          return;
        }
      }

      // 2. Add User Message
      const userMsgId = Date.now().toString();
      const userMessage: Message = {
        id: userMsgId,
        role: 'user',
        content,
        createdAt: Date.now(),
        status: 'completed'
      };

      // 3. Prepare Assistant Message
      const assistantMsgId = (Date.now() + 1).toString();
      const assistantMessage: Message = {
        id: assistantMsgId,
        role: 'assistant',
        content: '',
        createdAt: Date.now(),
        status: 'streaming',
        toolCalls: [],
        references: []
      };

      set((state) => ({
        messages: [...state.messages, userMessage, assistantMessage],
        isStreaming: true
      }));

      // 4. Start Streaming
      abortController = new AbortController();
      
      try {
        await fetchEventSource('/api/v1/chat/stream', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: JSON.stringify({
            session_id: sessionId,
            agent_id: agentId,
            message: content,
            stream: true,
          }),
          signal: abortController.signal,
          
          async onopen(response) {
            if (response.ok) {
              return; // everything's good
            } else {
              throw new Error(`Failed to send message: ${response.statusText}`);
            }
          },

          onmessage(msg) {
            if (msg.event === 'FatalError') {
              throw new Error(msg.data);
            }
            if (!msg.data) return;
            if (msg.data === '[DONE]') return;

            try {
              const event = JSON.parse(msg.data);
              
              set((state) => {
                const msgs = [...state.messages];
                const lastMsgIndex = msgs.findIndex(m => m.id === assistantMsgId);
                if (lastMsgIndex === -1) return state;

                const lastMsg = { ...msgs[lastMsgIndex] };

                switch (event.response_type) {
                  case 'answer':
                    lastMsg.content += (event.content || '');
                    break;
                  case 'thinking':
                    lastMsg.thinking = (lastMsg.thinking || '') + (event.content || '');
                    break;
                  case 'tool_call':
                    // WeKnora style: tool_calls are events. 
                    // Assuming event.tool_calls is a single tool call object or array
                    const newCalls = Array.isArray(event.tool_calls) ? event.tool_calls : [event.tool_calls];
                    lastMsg.toolCalls = [...(lastMsg.toolCalls || []), ...newCalls];
                    break;
                  case 'references':
                    lastMsg.references = event.data?.references || [];
                    break;
                  case 'stop':
                    lastMsg.status = 'completed';
                    break;
                  case 'error':
                    lastMsg.status = 'error';
                    lastMsg.error = event.error;
                    break;
                }
                
                msgs[lastMsgIndex] = lastMsg;
                return { messages: msgs };
              });
            } catch (e) {
              console.error('Error parsing SSE event:', e);
            }
          },

          onclose() {
            set((state) => {
                const msgs = [...state.messages];
                const lastMsgIndex = msgs.findIndex(m => m.id === assistantMsgId);
                if (lastMsgIndex !== -1 && msgs[lastMsgIndex].status === 'streaming') {
                    msgs[lastMsgIndex].status = 'completed';
                    return { messages: msgs, isStreaming: false };
                }
                return { isStreaming: false };
            });
          },

          onerror(err) {
            console.error('Stream error:', err);
             set((state) => {
                const msgs = [...state.messages];
                const lastMsgIndex = msgs.findIndex(m => m.id === assistantMsgId);
                if (lastMsgIndex !== -1) {
                    msgs[lastMsgIndex].status = 'error';
                    msgs[lastMsgIndex].error = err.message;
                }
                return { messages: msgs, isStreaming: false };
            });
            throw err; // rethrow to stop retries
          }
        });
      } catch (err: any) {
        // Error handling is mostly done in onerror
      } finally {
        // Update session title if it's new (optional, can refetch sessions)
        get().fetchSessions();
      }
    }
  };
});
