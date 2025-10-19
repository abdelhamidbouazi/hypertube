'use client'

import { useState } from 'react'
import { Toaster } from 'react-hot-toast'
import { ProfileCard, AccountCard, SecurityCard, ActivityCard } from '@/components/profile'

// Mock user data
const mockUser = {
  username: 'cinephile42',
  firstName: 'Alex',
  lastName: 'Johnson',
  bio: 'Passionate about cinema and storytelling. Love discovering hidden gems and discussing films with fellow movie enthusiasts.',
  email: 'alex.johnson@example.com',
  avatar: undefined as string | undefined,
  connectedAccounts: {
    fortyTwo: true,
    google: false
  }
}

const mockSessions = [
  {
    id: '1',
    device: 'MacBook Pro',
    location: 'San Francisco, CA',
    lastSeen: '2 hours ago',
    isCurrent: true
  },
  {
    id: '2',
    device: 'iPhone 15',
    location: 'San Francisco, CA',
    lastSeen: '1 day ago',
    isCurrent: false
  },
  {
    id: '3',
    device: 'Windows PC',
    location: 'New York, NY',
    lastSeen: '3 days ago',
    isCurrent: false
  }
]

const mockActivity = [
  {
    id: '1',
    type: 'watched' as const,
    title: 'The Matrix',
    timestamp: '2 hours ago',
    icon: '👁️'
  },
  {
    id: '2',
    type: 'commented' as const,
    title: 'Inception',
    timestamp: '1 day ago',
    icon: '💬'
  },
  {
    id: '3',
    type: 'favorited' as const,
    title: 'Blade Runner 2049',
    timestamp: '2 days ago',
    icon: '❤️'
  },
  {
    id: '4',
    type: 'watched' as const,
    title: 'Dune',
    timestamp: '3 days ago',
    icon: '👁️'
  }
]

const mockStats = {
  watchedCount: 42,
  commentsCount: 15
}

type User = {
  username: string
  firstName: string
  lastName: string
  bio: string
  email: string
  avatar: string | undefined
  connectedAccounts: {
    fortyTwo: boolean
    google: boolean
  }
}

export default function ProfilePage() {
  const [user, setUser] = useState<User>(mockUser)

  // Mock API functions with setTimeout to simulate network calls
  const handleProfileSave = async (data: { firstName: string; lastName: string; bio: string; avatar?: string }) => {
    return new Promise<void>((resolve) => {
      setTimeout(() => {
        setUser((prev: User) => ({
          ...prev,
          firstName: data.firstName,
          lastName: data.lastName,
          bio: data.bio,
          avatar: data.avatar
        }))
        resolve()
      }, 1000)
    })
  }

  const handleEmailChange = async (email: string) => {
    return new Promise<void>((resolve) => {
      setTimeout(() => {
        setUser((prev: User) => ({ ...prev, email }))
        resolve()
      }, 800)
    })
  }

  const handlePasswordChange = async (passwords: { current: string; new: string; confirm: string }) => {
    return new Promise<void>((resolve) => {
      setTimeout(() => {
        // Simulate password change
        resolve()
      }, 1200)
    })
  }

  const handleAccountConnect = async (provider: 'fortyTwo' | 'google') => {
    return new Promise<void>((resolve) => {
      setTimeout(() => {
        setUser((prev: User) => ({
          ...prev,
          connectedAccounts: {
            ...prev.connectedAccounts,
            [provider]: !prev.connectedAccounts[provider]
          }
        }))
        resolve()
      }, 600)
    })
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-background via-content1/30 to-background">
      <Toaster
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: 'hsl(var(--heroui-content1))',
            color: 'hsl(var(--heroui-foreground))',
            border: '1px solid hsl(var(--heroui-divider))',
            borderRadius: '12px',
            boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
          },
        }}
      />
      
      <div className="container mx-auto px-4 py-12 max-w-7xl">
        {/* Page Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-foreground mb-4">My Profile</h1>
          <p className="text-lg text-default-500 max-w-2xl mx-auto">
            Manage your account settings, security preferences, and track your movie journey
          </p>
        </div>

        {/* Two-column layout */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-8">
          {/* Left Column */}
          <div className="space-y-8 flex flex-col">
            <div className="flex-1">
              <ProfileCard
                user={user}
                onSave={handleProfileSave}
              />
            </div>
            
            <div className="flex-1">
              <AccountCard
                user={user}
                onEmailChange={handleEmailChange}
                onPasswordChange={handlePasswordChange}
                onAccountConnect={handleAccountConnect}
              />
            </div>
          </div>

          {/* Right Column */}
          <div className="space-y-8 flex flex-col">
            <div className="flex-1">
              <SecurityCard sessions={mockSessions} />
            </div>
            
            <div className="flex-1">
              <ActivityCard
                stats={mockStats}
                recentActivity={mockActivity}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}