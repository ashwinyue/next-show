import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import './index.css'
import { Layout } from './components/common/Layout'
import { LoginPage } from './pages/auth/LoginPage'
import { ChatPage } from './pages/chat/ChatPage'
import { AgentPage } from './pages/agent/AgentPage'
import { KnowledgePage } from './pages/knowledge/KnowledgePage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        index: true,
        element: <ChatPage />,
      },
      {
        path: 'agents',
        element: <AgentPage />,
      },
      {
        path: 'knowledge',
        element: <KnowledgePage />,
      },
      {
        path: 'settings',
        element: <div className="p-8"><h1 className="text-3xl font-bold font-mono text-primary">Settings</h1><p className="text-muted-foreground mt-4">Settings coming soon.</p></div>,
      },
    ],
  },
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
