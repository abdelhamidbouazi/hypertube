'use client'
import { useState } from 'react'
import clsx from 'clsx'
import { Navbar } from '@/components/navbar'
import { Sidebar } from '@/components/Sidebar'
import Footer from '@/components/footer'

interface LayoutWrapperProps {
  children: React.ReactNode
}

export const LayoutWrapper: React.FC<LayoutWrapperProps> = ({ children }) => {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  const handleSidebarCollapseToggle = () => {
    setSidebarCollapsed(!sidebarCollapsed)
  }

  return (
    <div className="relative min-h-screen overflow-x-hidden">
      <Sidebar
        isOpen={false}
        onToggle={() => {}}
        isCollapsed={sidebarCollapsed}
        onCollapseToggle={handleSidebarCollapseToggle}
      />
      <div className={clsx(
        "transition-all duration-300",
        sidebarCollapsed ? "lg:ml-16" : "lg:ml-56"
      )}>
        <Navbar
          onSidebarToggle={() => {}}
          onSidebarCollapseToggle={handleSidebarCollapseToggle}
          isSidebarCollapsed={sidebarCollapsed}
        />
        <main className="overflow-x-hidden">
          <div className="p-6">
            <div className="container mx-auto max-w-7xl">
              {children}
            </div>
            <Footer />
          </div>
        </main>
      </div>
    </div>
  )
}
