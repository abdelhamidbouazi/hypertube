'use client'

import { useState } from 'react'
import { Button } from '@heroui/button'
import { Card, CardBody, CardHeader } from '@heroui/card'
import { Input } from '@heroui/input'
import { Chip } from '@heroui/chip'
import { Divider } from '@heroui/divider'
import { toast } from 'react-hot-toast'

interface AccountCardProps {
  user: {
    email: string
    connectedAccounts: {
      fortyTwo: boolean
      google: boolean
    }
  }
  onEmailChange: (email: string) => Promise<void>
  onPasswordChange: (passwords: { current: string; new: string; confirm: string }) => Promise<void>
  onAccountConnect: (provider: 'fortyTwo' | 'google') => Promise<void>
}

export default function AccountCard({ user, onEmailChange, onPasswordChange, onAccountConnect }: AccountCardProps) {
  const [isEditingEmail, setIsEditingEmail] = useState(false)
  const [isChangingPassword, setIsChangingPassword] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [email, setEmail] = useState(user.email)
  const [passwords, setPasswords] = useState({
    current: '',
    new: '',
    confirm: ''
  })

  const handleEmailSave = async () => {
    if (!email || !email.includes('@')) {
      toast.error('Please enter a valid email address')
      return
    }

    setIsSaving(true)
    try {
      await onEmailChange(email)
      setIsEditingEmail(false)
      toast.success('Email updated successfully')
    } catch (error) {
      toast.error('Failed to update email')
    } finally {
      setIsSaving(false)
    }
  }

  const handlePasswordChange = async () => {
    if (!passwords.current || !passwords.new || !passwords.confirm) {
      toast.error('Please fill in all password fields')
      return
    }

    if (passwords.new !== passwords.confirm) {
      toast.error('New passwords do not match')
      return
    }

    if (passwords.new.length < 8) {
      toast.error('Password must be at least 8 characters')
      return
    }

    setIsSaving(true)
    try {
      await onPasswordChange(passwords)
      setPasswords({ current: '', new: '', confirm: '' })
      setIsChangingPassword(false)
      toast.success('Password changed successfully')
    } catch (error) {
      toast.error('Failed to change password')
    } finally {
      setIsSaving(false)
    }
  }

  const handleAccountConnect = async (provider: 'fortyTwo' | 'google') => {
    try {
      await onAccountConnect(provider)
      toast.success(`${provider === 'fortyTwo' ? '42' : 'Google'} account ${user.connectedAccounts[provider] ? 'disconnected' : 'connected'}`)
    } catch (error) {
      toast.error(`Failed to ${user.connectedAccounts[provider] ? 'disconnect' : 'connect'} ${provider === 'fortyTwo' ? '42' : 'Google'} account`)
    }
  }

  return (
    <Card className="shadow-lg border-0 bg-gradient-to-br from-content1 to-content2 h-full">
      <CardHeader className="pb-4">
        <div>
          <h2 className="text-xl font-bold text-foreground">Account Settings</h2>
          <p className="text-sm text-default-500 mt-1">Manage your account security and connected services</p>
        </div>
      </CardHeader>
      
      <Divider className="mb-6" />
      
      <CardBody className="space-y-8">
        {/* Email Section */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-base font-semibold text-foreground">Email Address</h3>
              <p className="text-sm text-default-500">Your primary email for account notifications</p>
            </div>
            {!isEditingEmail && (
              <Button 
                variant="flat" 
                size="sm"
                onPress={() => setIsEditingEmail(true)}
                className="font-medium"
              >
                Edit
              </Button>
            )}
          </div>
          
          <div className="space-y-3">
            <Input
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              variant="bordered"
              isDisabled={!isEditingEmail}
              placeholder="Enter your email address"
              type="email"
            />
            
            {isEditingEmail && (
              <div className="flex gap-3">
                <Button
                  color="primary"
                  size="sm"
                  onPress={handleEmailSave}
                  isLoading={isSaving}
                  isDisabled={isSaving}
                  className="font-medium"
                >
                  Save Email
                </Button>
                <Button
                  variant="light"
                  size="sm"
                  onPress={() => {
                    setEmail(user.email)
                    setIsEditingEmail(false)
                  }}
                  isDisabled={isSaving}
                  className="font-medium"
                >
                  Cancel
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* Password Section */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-base font-semibold text-foreground">Password</h3>
              <p className="text-sm text-default-500">Update your account password</p>
            </div>
            <Button
              variant="flat"
              size="sm"
              onPress={() => setIsChangingPassword(!isChangingPassword)}
              className="font-medium"
            >
              {isChangingPassword ? 'Cancel' : 'Change Password'}
            </Button>
          </div>

          {isChangingPassword && (
            <div className="space-y-4 p-4 bg-content2/30 rounded-lg border border-divider">
              <div className="space-y-3">
                <Input
                  type="password"
                  placeholder="Current password"
                  value={passwords.current}
                  onChange={(e) => setPasswords(prev => ({ ...prev, current: e.target.value }))}
                  variant="bordered"
                />
                <Input
                  type="password"
                  placeholder="New password"
                  value={passwords.new}
                  onChange={(e) => setPasswords(prev => ({ ...prev, new: e.target.value }))}
                  variant="bordered"
                />
                <Input
                  type="password"
                  placeholder="Confirm new password"
                  value={passwords.confirm}
                  onChange={(e) => setPasswords(prev => ({ ...prev, confirm: e.target.value }))}
                  variant="bordered"
                />
              </div>
              <Button
                color="primary"
                onPress={handlePasswordChange}
                isLoading={isSaving}
                isDisabled={isSaving}
                className="w-full font-medium"
              >
                {isSaving ? 'Updating...' : 'Update Password'}
              </Button>
            </div>
          )}
        </div>

        {/* Connected Accounts */}
        <div className="space-y-4">
          <div>
            <h3 className="text-base font-semibold text-foreground">Connected Accounts</h3>
            <p className="text-sm text-default-500">Link your accounts for easier sign-in</p>
          </div>
          
          <div className="space-y-3">
            {/* 42 Account */}
            <div className="flex items-center justify-between p-4 rounded-lg border border-divider bg-content2/20">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
                  <span className="text-primary font-bold text-sm">42</span>
                </div>
                <div>
                  <p className="font-medium text-foreground">42 School</p>
                  <div className="flex items-center gap-2">
                    <Chip
                      size="sm"
                      color={user.connectedAccounts.fortyTwo ? 'success' : 'default'}
                      variant="flat"
                    >
                      {user.connectedAccounts.fortyTwo ? 'Connected' : 'Not connected'}
                    </Chip>
                  </div>
                </div>
              </div>
              <Button
                variant="flat"
                size="sm"
                color={user.connectedAccounts.fortyTwo ? 'danger' : 'primary'}
                onPress={() => handleAccountConnect('fortyTwo')}
                className="font-medium"
              >
                {user.connectedAccounts.fortyTwo ? 'Disconnect' : 'Connect'}
              </Button>
            </div>

            {/* Google Account */}
            <div className="flex items-center justify-between p-4 rounded-lg border border-divider bg-content2/20">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-danger/10 rounded-lg flex items-center justify-center">
                  <span className="text-danger font-bold text-sm">G</span>
                </div>
                <div>
                  <p className="font-medium text-foreground">Google</p>
                  <div className="flex items-center gap-2">
                    <Chip
                      size="sm"
                      color={user.connectedAccounts.google ? 'success' : 'default'}
                      variant="flat"
                    >
                      {user.connectedAccounts.google ? 'Connected' : 'Not connected'}
                    </Chip>
                  </div>
                </div>
              </div>
              <Button
                variant="flat"
                size="sm"
                color={user.connectedAccounts.google ? 'danger' : 'primary'}
                onPress={() => handleAccountConnect('google')}
                className="font-medium"
              >
                {user.connectedAccounts.google ? 'Disconnect' : 'Connect'}
              </Button>
            </div>
          </div>
          
          <p className="text-xs text-default-400">
            Connect your accounts to sign in faster and sync your preferences across platforms.
          </p>
        </div>
      </CardBody>
    </Card>
  )
}
