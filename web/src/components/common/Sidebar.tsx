import { Link, useLocation } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  MessageSquare,
  Bot,
  Database,
  Settings,
  LogOut,
  Menu,
  X
} from 'lucide-react';
import { useAuthStore } from '@/stores/useAuthStore';
import { useState } from 'react';

const sidebarItems = [
  { icon: MessageSquare, label: 'Chat', href: '/' },
  { icon: Bot, label: 'Agents', href: '/agents' },
  { icon: Database, label: 'Knowledge', href: '/knowledge' },
  { icon: Settings, label: 'Settings', href: '/settings' },
];

export function Sidebar() {
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const [isOpen, setIsOpen] = useState(true);

  return (
    <>
      <div className={cn(
        "fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r bg-card transition-all duration-300 ease-in-out lg:static",
        !isOpen && "lg:w-16 -translate-x-full lg:translate-x-0"
      )}>
        <div className={cn("flex h-14 items-center border-b px-4", isOpen ? "justify-between" : "justify-center")}>
          <Link to="/" className={cn("flex items-center gap-2 font-bold", !isOpen && "hidden")}>
            <span className="text-xl text-primary">NextShow</span>
          </Link>
          <Button variant="ghost" size="icon" onClick={() => setIsOpen(!isOpen)} className="hidden lg:flex">
             {isOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </Button>
          <Button variant="ghost" size="icon" onClick={() => setIsOpen(!isOpen)} className="lg:hidden absolute right-[-40px] top-2 bg-background border rounded-md">
            {isOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </Button>
        </div>

        <nav className="flex-1 space-y-2 p-2">
          {sidebarItems.map((item) => {
            const isActive = location.pathname === item.href;
            return (
              <Link
                key={item.href}
                to={item.href}
                className={cn(
                  "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground",
                  isActive && "bg-accent text-accent-foreground",
                  !isOpen && "justify-center px-2"
                )}
                title={!isOpen ? item.label : undefined}
              >
                <item.icon className="h-5 w-5" />
                {isOpen && <span>{item.label}</span>}
              </Link>
            );
          })}
        </nav>

        <div className="border-t p-4">
          <div className={cn("flex items-center gap-3", !isOpen && "justify-center")}>
            {isOpen ? (
              <>
                <div className="flex flex-1 flex-col overflow-hidden">
                  <span className="truncate text-sm font-medium">{user?.username || 'User'}</span>
                  <span className="truncate text-xs text-muted-foreground">{user?.email}</span>
                </div>
                <Button variant="ghost" size="icon" onClick={logout}>
                  <LogOut className="h-5 w-5" />
                </Button>
              </>
            ) : (
               <Button variant="ghost" size="icon" onClick={logout} title="Logout">
                  <LogOut className="h-5 w-5" />
                </Button>
            )}
          </div>
        </div>
      </div>
      {/* Mobile Overlay */}
      {isOpen && (
        <div 
          className="fixed inset-0 z-40 bg-background/80 backdrop-blur-sm lg:hidden"
          onClick={() => setIsOpen(false)}
        />
      )}
    </>
  );
}
