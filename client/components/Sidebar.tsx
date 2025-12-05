'use client'
import React, { useState, useEffect } from 'react'
import NextLink from 'next/link'
import { usePathname } from 'next/navigation'
import clsx from 'clsx'
import { Button } from '@heroui/button'
import { Avatar } from '@heroui/avatar'
import { Divider } from '@heroui/divider'
import {
  Dropdown,
  DropdownTrigger,
  DropdownMenu,
  DropdownItem,
} from '@heroui/dropdown'
import { logoutUser, useMe, useUserStats, useWatchHistory } from '@/lib/hooks'
import { ThemeSwitch } from '@/components/theme-switch'
import { useTheme } from 'next-themes'
import {
  Logo,
  GithubIcon,
  TwitterIcon,
  DiscordIcon,
} from '@/components/icons'
import {
  Menu,
  X,
  Home,
  Compass,
  Settings,
  User,
  Bookmark,
  History,
  LogOut,
  ChevronLeft,
  ChevronRight,
  Film,
  Star,
  Clock,
  TrendingUp,
  Sun,
  Moon,
} from 'lucide-react'

interface SidebarProps {
  isOpen: boolean
  onToggle: () => void
  isCollapsed: boolean
  onCollapseToggle: () => void
}

export const Sidebar: React.FC<SidebarProps> = ({
  isOpen,
  onToggle,
  isCollapsed,
  onCollapseToggle,
}) => {
  const pathname = usePathname()
  const { theme, setTheme } = useTheme()
  const { user, isLoading: userLoading } = useMe()
  const { stats, isLoading: statsLoading } = useUserStats()
  const { continueWatching, isLoading: watchHistoryLoading } = useWatchHistory()

  // main navigation items
  const navigation = [
    { name: 'Discover', href: '/app/discover', icon: Compass },
    { name: 'Watch Later', href: '/app/watch-later', icon: Clock },
    { name: 'Watch History', href: '/app/history', icon: History },
    { name: 'Settings', href: '/app/settings', icon: Settings },
  ]

  // user dropdown menu items
  const userMenuItems = [
    { key: 'profile', label: 'My Profile', icon: User, href: '/app/profile' },
  ]

  // check if current route is active
  const isActive = (href: string) => pathname === href

  // toggle between light and dark theme
  const toggleTheme = () => {
    // setTheme(theme === 'light' ? 'dark' : 'light')
    setTheme(theme === 'light' ? 'dark' : 'dark')
  }

  // close sidebar on mobile route change
  useEffect(() => {
    if (window.innerWidth < 1024) {
      onToggle()
    }
  }, [pathname])

  return (
    <>
      {/* mobile overlay for sidebar */}
      {isOpen && (
        <div 
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={onToggle}
        />
      )}

      {/* main sidebar container */}
      <aside className={clsx(
        'fixed left-0 top-0 z-50 flex flex-col bg-content1 border-r border-divider transition-all duration-300 ease-in-out min-h-screen',
        isOpen ? 'translate-x-0' : '-translate-x-full',
        'lg:translate-x-0',
        isCollapsed ? 'lg:w-16' : 'lg:w-56'
      )}>
        {/* sidebar header with logo */}
        <div className="flex items-center justify-between p-3 border-b border-divider">
          {!isCollapsed && (
            <NextLink href="/app/discover" className="flex items-center gap-2">
              <Logo />
              <span className="text-lg font-light text-foreground">CINÉTHOS</span>
            </NextLink>
          )}
          {isCollapsed && (
            <NextLink href="/app/discover" className="flex justify-center w-full">
              <Logo />
            </NextLink>
          )}
          
        </div>

        {/* main navigation menu */}
        <nav className="flex-1 px-3 py-4 space-y-1">
          {navigation.map((item) => {
            const Icon = item.icon
            const active = isActive(item.href)
            
            return (
              <NextLink
                key={item.name}
                href={item.href}
                className={clsx(
                  'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200',
                  active
                    ? 'bg-primary text-primary-foreground shadow-sm'
                    : 'text-foreground hover:bg-content2 hover:text-primary',
                  isCollapsed && 'justify-center px-2'
                )}
              >
                <Icon className={clsx(
                  'w-5 h-5 flex-shrink-0 transition-transform',
                  active && 'scale-105'
                )} />
                {!isCollapsed && (
                  <span className="transition-opacity duration-200">
                    {item.name}
                  </span>
                )}
              </NextLink>
            )
          })}
        </nav>

        <Divider />

        {/* continue watching section */}
        {!isCollapsed && (
          <div className="px-3 py-4">
            <h3 className="text-xs font-semibold text-default-500 uppercase tracking-wider mb-3">
              Continue Watching
            </h3>
            {watchHistoryLoading ? (
              <div className="bg-gradient-to-br from-primary/10 to-secondary/10 rounded-xl p-3 border border-primary/20">
                <div className="flex items-center gap-3">
                  <div className="w-12 h-16 bg-content2 rounded-lg animate-pulse" />
                  <div className="flex-1 min-w-0">
                    <div className="h-4 bg-content2 rounded animate-pulse mb-2" />
                    <div className="h-3 bg-content2 rounded animate-pulse w-2/3" />
                  </div>
                </div>
              </div>
            ) : continueWatching ? (
              <div className="bg-gradient-to-br from-primary/10 to-secondary/10 rounded-xl p-3 border border-primary/20">
                <div className="flex items-center gap-3">
                  <div className="relative">
                    <img 
                      src={continueWatching.posterPath || "/placeholder-poster.jpg"} 
                      alt={`${continueWatching.title} poster`}
                      className="w-12 h-16 object-cover rounded-lg"
                    />
                  </div>
                  <div className="flex-1 min-w-0">
                    <h4 className="text-sm font-semibold text-foreground truncate">
                      {continueWatching.title}
                    </h4>
                    <p className="text-xs text-default-500 truncate">
                      {continueWatching.genre} • {continueWatching.year}
                    </p>
                  </div>
                </div>
                <div className="mt-3 flex gap-2">
                  <Button
                    size="sm"
                    color="success"
                    variant="flat"
                    className="flex-1"
                    as={NextLink}
                    href={`/app/movie/${continueWatching.id}/watch`}
                  >
                    Resume
                  </Button>
                  <Button
                    size="sm"
                    variant="flat"
                    color="danger"
                    isIconOnly
                    className="text-default-500"
                  >
                    <X className="w-3 h-3" />
                  </Button>
                </div>
              </div>
            ) : (
              <div className="bg-gradient-to-br from-content2/50 to-content2/30 rounded-xl p-3 border border-divider">
                <p className="text-sm text-default-500 text-center">
                  No movies in progress
                </p>
              </div>
            )}
          </div>
        )}

        <Divider />

        {/* user stats section */}
        {!isCollapsed && (
          <div className="px-3 py-4">
            <h3 className="text-xs font-semibold text-default-500 uppercase tracking-wider mb-3">
              Quick Stats
            </h3>
            <div className="space-y-3">
              <div className="flex justify-between items-center text-sm p-2 rounded-lg bg-content2/50">
                <span className="text-default-500">Movies Watched</span>
                <span className="font-semibold text-primary">
                  {statsLoading ? '...' : stats.moviesWatched}
                </span>
              </div>
              <div className="flex justify-between items-center text-sm p-2 rounded-lg bg-content2/50">
                <span className="text-default-500">Hours</span>
                <span className="font-semibold text-primary">
                  {statsLoading ? '...' : stats.hoursWatched}
                </span>
              </div>
              {/* <div className="flex justify-between items-center text-sm p-2 rounded-lg bg-content2/50">
                <span className="text-default-500">Favorites</span>
                <span className="font-semibold text-primary">
                  {statsLoading ? '...' : stats.favorites}
                </span>
              </div> */}
            </div>
          </div>
        )}

        <Divider />

        {/* user profile dropdown */}
        <div className="p-3">
          <Dropdown placement="top-start">
            <DropdownTrigger>
              <Button
                variant="flat"
                className={clsx(
                  'w-full justify-start gap-3 p-3 h-auto rounded-xl transition-all duration-200 hover:bg-content2',
                  isCollapsed && 'justify-center px-2'
                )}
              >
                <Avatar
                  size="sm"
                  src={user?.avatar || "https://i.pravatar.cc/150?u=user"}
                  className="flex-shrink-0 ring-2 ring-primary/20"
                />
                {!isCollapsed && (
                  <div className="flex flex-col items-start text-left">
                    <span className="text-sm font-semibold">
                      {userLoading ? 'Loading...' : user ? `${user.firstname} ${user.lastname}` : 'User Name'}
                    </span>
                    <span className="text-xs text-default-500">
                      {userLoading ? 'Loading...' : user?.email || 'user@example.com'}
                    </span>
                  </div>
                )}
              </Button>
            </DropdownTrigger>
            <DropdownMenu aria-label="User menu" className="min-w-[200px]">
              <DropdownItem
                key="profile"
                startContent={<User className="w-4 h-4" />}
                as={NextLink}
                href="/app/profile"
                className="rounded-lg"
              >
                My Profile
              </DropdownItem>
              <DropdownItem
                key="theme"
                startContent={theme === 'light' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
                onPress={toggleTheme}
                className="rounded-lg"
              >
                {theme === 'light' ? 'Dark Mode' : 'Light Mode'}
              </DropdownItem>
              <DropdownItem
                key="logout"
                color="danger"
                startContent={<LogOut className="w-4 h-4" />}
                onPress={() => logoutUser()}
                className="rounded-lg"
              >
                Log Out
              </DropdownItem>
            </DropdownMenu>
          </Dropdown>
        </div>

        {/* social media links */}
        <div className="px-3 pb-4">
          <div className="flex gap-1 justify-center">
            <Button
              isIconOnly
              variant="flat"
              size="sm"
              as={NextLink}
              href="https://x.com/abdelhamidbouazi"
              target="_blank"
              className="hover:bg-content2 transition-colors duration-200"
            >
              <TwitterIcon className="w-4 h-4" />
            </Button>
            <Button
              isIconOnly
              variant="flat"
              size="sm"
              as={NextLink}
              href="https://abouazi.me"
              target="_blank"
              className="hover:bg-content2 transition-colors duration-200"
            >
              <DiscordIcon className="w-4 h-4" />
            </Button>
            <Button
              isIconOnly
              variant="flat"
              size="sm"
              as={NextLink}
              href="https://github.com/abdelhamidbouazi"
              target="_blank"
              className="hover:bg-content2 transition-colors duration-200"
            >
              <GithubIcon className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </aside>
    </>
  )
}