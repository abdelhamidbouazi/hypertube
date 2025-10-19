'use client'

import { Card, CardBody, CardHeader } from '@heroui/card'
import { Chip } from '@heroui/chip'
import { Divider } from '@heroui/divider'
import { Button } from '@heroui/button'

interface SecurityCardProps {
  sessions: Array<{
    id: string
    device: string
    location: string
    lastSeen: string
    isCurrent: boolean
  }>
}

export default function SecurityCard({ sessions }: SecurityCardProps) {
  const getDeviceIcon = (device: string) => {
    if (device.toLowerCase().includes('iphone') || device.toLowerCase().includes('mobile')) {
      return '📱'
    } else if (device.toLowerCase().includes('macbook') || device.toLowerCase().includes('mac')) {
      return '💻'
    } else if (device.toLowerCase().includes('windows') || device.toLowerCase().includes('pc')) {
      return '🖥️'
    } else {
      return '📱'
    }
  }

  return (
    <Card className="shadow-lg border-0 bg-gradient-to-br from-content1 to-content2 h-full">
      <CardHeader className="pb-4">
        <div>
          <h2 className="text-xl font-bold text-foreground">Security & Privacy</h2>
          <p className="text-sm text-default-500 mt-1">Manage your account security and active sessions</p>
        </div>
      </CardHeader>
      
      <Divider className="mb-6" />
      
      <CardBody className="space-y-6">
        {/* Privacy Notice */}
        <div className="p-4 bg-success/10 rounded-lg border border-success/20">
          <div className="flex items-start gap-3">
            <div className="w-6 h-6 bg-success rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
              <span className="text-white text-xs">✓</span>
            </div>
            <div>
              <p className="text-sm font-medium text-success">
                Your email is private
              </p>
              <p className="text-xs text-success/80 mt-1">
                Other users cannot see your email address. Only you can view and manage it.
              </p>
            </div>
          </div>
        </div>

        {/* Active Sessions */}
        <div className="space-y-4">
          <div>
            <h3 className="text-base font-semibold text-foreground">Active Sessions</h3>
            <p className="text-sm text-default-500">Devices currently signed into your account</p>
          </div>
          
          <div className="space-y-3">
            {sessions.map((session) => (
              <div
                key={session.id}
                className={`flex items-center justify-between p-4 rounded-lg border transition-colors ${
                  session.isCurrent 
                    ? 'border-primary/30 bg-primary/5' 
                    : 'border-divider bg-content2/20 hover:bg-content2/30'
                }`}
              >
                <div className="flex items-center gap-3">
                  <div className="text-2xl">{getDeviceIcon(session.device)}</div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-foreground">{session.device}</p>
                      {session.isCurrent && (
                        <Chip size="sm" color="primary" variant="flat">
                          Current
                        </Chip>
                      )}
                    </div>
                    <p className="text-sm text-default-500">{session.location}</p>
                    <p className="text-xs text-default-400">Last seen: {session.lastSeen}</p>
                  </div>
                </div>
                {!session.isCurrent && (
                  <Button
                    size="sm"
                    variant="light"
                    color="danger"
                    className="font-medium"
                  >
                    Revoke
                  </Button>
                )}
              </div>
            ))}
          </div>
          
          <p className="text-xs text-default-400">
            You can revoke access for any device you no longer use or recognize.
          </p>
        </div>
      </CardBody>
    </Card>
  )
}
