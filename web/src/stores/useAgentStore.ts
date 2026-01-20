import { create } from 'zustand';
import { api } from '@/api/client';

export interface Agent {
  id: string;
  name: string;
  display_name: string;
  description: string;
  agent_type: string;
  agent_role: 'orchestrator' | 'specialist';
  model_name: string;
  is_enabled: boolean;
  created_at: string;
}

interface AgentState {
  agents: Agent[];
  selectedAgentId: string | null;
  isLoading: boolean;
  
  // Actions
  fetchAgents: () => Promise<void>;
  createAgent: (data: Partial<Agent>) => Promise<Agent>;
  updateAgent: (id: string, data: Partial<Agent>) => Promise<Agent>;
  deleteAgent: (id: string) => Promise<void>;
  selectAgent: (id: string | null) => void;
}

export const useAgentStore = create<AgentState>((set) => ({
  agents: [],
  selectedAgentId: null,
  isLoading: false,

  fetchAgents: async () => {
    set({ isLoading: true });
    try {
      const response = await api.get('/agents');
      set({ agents: response.data.agents || [] });
    } catch (error) {
      console.error('Failed to fetch agents:', error);
    } finally {
      set({ isLoading: false });
    }
  },

  createAgent: async (data) => {
    try {
      const response = await api.post('/agents', data);
      const newAgent = response.data;
      set((state) => ({ agents: [newAgent, ...state.agents] }));
      return newAgent;
    } catch (error) {
      console.error('Failed to create agent:', error);
      throw error;
    }
  },

  updateAgent: async (id, data) => {
    try {
      const response = await api.put(`/agents/${id}`, data);
      const updatedAgent = response.data;
      set((state) => ({
        agents: state.agents.map((a) => (a.id === id ? updatedAgent : a)),
      }));
      return updatedAgent;
    } catch (error) {
      console.error('Failed to update agent:', error);
      throw error;
    }
  },

  deleteAgent: async (id) => {
    try {
      await api.delete(`/agents/${id}`);
      set((state) => ({
        agents: state.agents.filter((a) => a.id !== id),
        selectedAgentId: state.selectedAgentId === id ? null : state.selectedAgentId,
      }));
    } catch (error) {
      console.error('Failed to delete agent:', error);
      throw error;
    }
  },

  selectAgent: (id) => set({ selectedAgentId: id }),
}));
