'use client'

import { Card, CardBody, CardHeader } from '@heroui/card'
import { Chip } from '@heroui/chip'
import { Divider } from '@heroui/divider'

interface ActivityCardProps {
  stats: {
    watchedCount: number
    commentsCount: number
  }
  recentActivity: Array<{
    id: string
    type: 'watched' | 'commented' | 'favorited'
    title: string
    timestamp: string
    icon: string
  }>
}

export default function ActivityCard({ stats, recentActivity }: ActivityCardProps) {
  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'watched':
        return '👁️'
      case 'commented':
        return '💬'
      case 'favorited':
        return '❤️'
      default:
        return '📝'
    }
  }

  const getActivityText = (type: string, title: string) => {
    switch (type) {
      case 'watched':
        return `Watched "${title}"`
      case 'commented':
        return `Commented on "${title}"`
      case 'favorited':
        return `Added "${title}" to favorites`
      default:
        return title
    }
  }

  const getActivityColor = (type: string) => {
    switch (type) {
      case 'watched':
        return 'primary'
      case 'commented':
        return 'secondary'
      case 'favorited':
        return 'danger'
      default:
        return 'default'
    }
  }

  return (
    <Card className="shadow-lg border-0 bg-gradient-to-br from-content1 to-content2 h-full">
      <CardHeader className="pb-4">
        <div>
          <h2 className="text-xl font-bold text-foreground">Activity Summary</h2>
          <p className="text-sm text-default-500 mt-1">Your movie watching and engagement stats</p>
        </div>
      </CardHeader>
      
      <Divider className="mb-6" />
      
      <CardBody className="space-y-6">
        {/* Stats */}
        <div className="space-y-4">
          <div>
            <h3 className="text-base font-semibold text-foreground">Your Stats</h3>
            <p className="text-sm text-default-500">Track your movie journey</p>
          </div>
          
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 rounded-lg border border-primary/20 bg-primary/5">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-primary/20 rounded-lg flex items-center justify-center">
                  <span className="text-primary text-lg">🎬</span>
                </div>
                <div>
                  <p className="text-2xl font-bold text-primary">{stats.watchedCount}</p>
                  <p className="text-xs text-primary/80 font-medium">Movies Watched</p>
                </div>
              </div>
            </div>
            
            <div className="p-4 rounded-lg border border-secondary/20 bg-secondary/5">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-secondary/20 rounded-lg flex items-center justify-center">
                  <span className="text-secondary text-lg">💬</span>
                </div>
                <div>
                  <p className="text-2xl font-bold text-secondary">{stats.commentsCount}</p>
                  <p className="text-xs text-secondary/80 font-medium">Comments Made</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Recent Activity */}
        <div className="space-y-4">
          <div>
            <h3 className="text-base font-semibold text-foreground">Recent Activity</h3>
            <p className="text-sm text-default-500">Your latest movie interactions</p>
          </div>
          
          <div className="space-y-3">
            {recentActivity.length > 0 ? (
              recentActivity.map((activity) => (
                <div
                  key={activity.id}
                  className="flex items-center gap-3 p-3 rounded-lg border border-divider bg-content2/20 hover:bg-content2/30 transition-colors"
                >
                  <div className="flex-shrink-0">
                    <div className={`w-8 h-8 rounded-lg flex items-center justify-center bg-${getActivityColor(activity.type)}/10`}>
                      <span className="text-lg">{getActivityIcon(activity.type)}</span>
                    </div>
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-foreground truncate">
                      {getActivityText(activity.type, activity.title)}
                    </p>
                    <p className="text-xs text-default-500">{activity.timestamp}</p>
                  </div>
                  <Chip
                    size="sm"
                    color={getActivityColor(activity.type)}
                    variant="flat"
                    className="flex-shrink-0"
                  >
                    {activity.type}
                  </Chip>
                </div>
              ))
            ) : (
              <div className="text-center py-8">
                <div className="w-16 h-16 bg-default/10 rounded-full flex items-center justify-center mx-auto mb-4">
                  <span className="text-2xl">🎬</span>
                </div>
                <p className="text-sm font-medium text-default-500">No recent activity</p>
                <p className="text-xs text-default-400 mt-1">
                  Start watching movies to see your activity here
                </p>
              </div>
            )}
          </div>
        </div>
      </CardBody>
    </Card>
  )
}
