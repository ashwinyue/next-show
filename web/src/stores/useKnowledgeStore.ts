import { create } from 'zustand';
import { api } from '@/api/client';

export interface KnowledgeBase {
  id: string;
  name: string;
  description: string;
  embedding_model: string;
  created_at: string;
}

export interface KnowledgeDocument {
  id: string;
  knowledge_base_id: string;
  title: string;
  content_type: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  error?: string;
  chunk_count: number;
  metadata: any;
  created_at: string;
}

export interface Chunk {
  id: string;
  document_id: string;
  content: string;
  token_count: number;
  index: number;
}

interface KnowledgeState {
  knowledgeBases: KnowledgeBase[];
  selectedKb: KnowledgeBase | null;
  documents: KnowledgeDocument[];
  chunks: Chunk[];
  isLoading: boolean;

  fetchKnowledgeBases: () => Promise<void>;
  createKnowledgeBase: (data: Partial<KnowledgeBase>) => Promise<KnowledgeBase>;
  deleteKnowledgeBase: (id: string) => Promise<void>;
  selectKnowledgeBase: (kb: KnowledgeBase | null) => void;
  
  fetchDocuments: (kbId: string) => Promise<void>;
  uploadDocument: (kbId: string, file: File, title?: string) => Promise<void>;
  deleteDocument: (id: string) => Promise<void>;
  
  fetchChunks: (docId: string) => Promise<void>;
}

export const useKnowledgeStore = create<KnowledgeState>((set, get) => ({
  knowledgeBases: [],
  selectedKb: null,
  documents: [],
  chunks: [],
  isLoading: false,

  fetchKnowledgeBases: async () => {
    set({ isLoading: true });
    try {
      const response = await api.get('/knowledge');
      set({ knowledgeBases: response.data.items || [] });
    } catch (error) {
      console.error('Failed to fetch KBs:', error);
    } finally {
      set({ isLoading: false });
    }
  },

  createKnowledgeBase: async (data) => {
    try {
      const response = await api.post('/knowledge', data);
      const newKb = response.data;
      set((state) => ({ knowledgeBases: [newKb, ...state.knowledgeBases] }));
      return newKb;
    } catch (error) {
      console.error('Failed to create KB:', error);
      throw error;
    }
  },

  deleteKnowledgeBase: async (id) => {
    try {
      await api.delete(`/knowledge/${id}`);
      set((state) => ({
        knowledgeBases: state.knowledgeBases.filter((kb) => kb.id !== id),
        selectedKb: state.selectedKb?.id === id ? null : state.selectedKb,
      }));
    } catch (error) {
      console.error('Failed to delete KB:', error);
      throw error;
    }
  },

  selectKnowledgeBase: (kb) => {
    set({ selectedKb: kb, documents: [], chunks: [] });
    if (kb) {
      get().fetchDocuments(kb.id);
    }
  },

  fetchDocuments: async (kbId) => {
    set({ isLoading: true });
    try {
      const response = await api.get(`/knowledge/${kbId}/documents`);
      set({ documents: response.data.items || [] });
    } catch (error) {
      console.error('Failed to fetch documents:', error);
    } finally {
      set({ isLoading: false });
    }
  },

  uploadDocument: async (kbId, file, title) => {
    const formData = new FormData();
    formData.append('file', file);
    if (title) formData.append('title', title);
    
    // Default chunk settings
    formData.append('chunk_size', '512');
    formData.append('chunk_overlap', '50');

    try {
      await api.post(`/knowledge/${kbId}/upload`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      await get().fetchDocuments(kbId);
    } catch (error) {
      console.error('Failed to upload document:', error);
      throw error;
    }
  },

  deleteDocument: async (id) => {
    try {
      await api.delete(`/documents/${id}`);
      set((state) => ({
        documents: state.documents.filter((d) => d.id !== id),
      }));
    } catch (error) {
      console.error('Failed to delete document:', error);
      throw error;
    }
  },

  fetchChunks: async (docId) => {
    set({ isLoading: true });
    try {
      const response = await api.get(`/documents/${docId}/chunks`);
      set({ chunks: response.data.items || [] });
    } catch (error) {
      console.error('Failed to fetch chunks:', error);
    } finally {
      set({ isLoading: false });
    }
  },
}));
