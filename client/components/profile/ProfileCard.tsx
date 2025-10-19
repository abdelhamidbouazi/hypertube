'use client'

import { useState } from 'react'
import { Avatar } from '@heroui/avatar'
import { Button } from '@heroui/button'
import { Card, CardBody, CardHeader } from '@heroui/card'
import { Input } from '@heroui/input'
import { Textarea } from '@heroui/input'
import { Divider } from '@heroui/divider'
import { toast } from 'react-hot-toast'
import AvatarUploader from './AvatarUploader'

interface ProfileCardProps {
  user: {
    username: string
    firstName: string
    lastName: string
    bio: string
    avatar?: string
  }
  onSave: (data: { firstName: string; lastName: string; bio: string; avatar?: string }) => Promise<void>
}

export default function ProfileCard({ user, onSave }: ProfileCardProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [formData, setFormData] = useState({
    firstName: user.firstName,
    lastName: user.lastName,
    bio: user.bio,
    avatar: user.avatar
  })
  const [originalData] = useState(formData)

  const handleSave = async () => {
    setIsSaving(true)
    try {
      await onSave(formData)
      setIsEditing(false)
      toast.success('Profile updated successfully')
    } catch (error) {
      toast.error('Failed to update profile')
    } finally {
      setIsSaving(false)
    }
  }

  const handleCancel = () => {
    setFormData(originalData)
    setIsEditing(false)
  }

  const handleAvatarChange = (file: File | null) => {
    if (file) {
      setFormData(prev => ({ ...prev, avatar: URL.createObjectURL(file) }))
    }
  }

  const handleAvatarRemove = () => {
    setFormData(prev => ({ ...prev, avatar: undefined }))
  }

  return (
    <Card className="shadow-lg border-0 bg-gradient-to-br from-content1 to-content2 h-full">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between w-full">
          <div>
            <h2 className="text-xl font-bold text-foreground">Personal Information</h2>
            <p className="text-sm text-default-500 mt-1">Manage your profile details and preferences</p>
          </div>
          {!isEditing && (
            <Button 
              color="primary" 
              variant="flat"
              onPress={() => setIsEditing(true)}
              className="font-medium"
            >
              Edit Profile
            </Button>
          )}
        </div>
      </CardHeader>
      
      <Divider className="mb-6" />
      
      <CardBody className="space-y-8">
        {/* Avatar Section */}
        <div className="flex flex-col lg:flex-row gap-8 items-start">
          <div className="flex-shrink-0">
            <AvatarUploader
              src={formData.avatar}
              name={`${formData.firstName} ${formData.lastName}`}
              onAvatarChange={handleAvatarChange}
              onRemove={handleAvatarRemove}
            />
          </div>
          
          {/* Form Fields */}
          <div className="flex-1 space-y-6">
            {/* Username - Read Only */}
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">Username</label>
              <Input
                value={user.username}
                isDisabled
                variant="bordered"
                className="bg-content2/50"
                classNames={{
                  input: "text-default-500"
                }}
              />
              <p className="text-xs text-default-400">Your username cannot be changed</p>
            </div>
            
            {/* Name Fields */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground">First Name</label>
                <Input
                  value={formData.firstName}
                  onChange={(e) => setFormData(prev => ({ ...prev, firstName: e.target.value }))}
                  variant="bordered"
                  isDisabled={!isEditing}
                  placeholder="Enter your first name"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground">Last Name</label>
                <Input
                  value={formData.lastName}
                  onChange={(e) => setFormData(prev => ({ ...prev, lastName: e.target.value }))}
                  variant="bordered"
                  isDisabled={!isEditing}
                  placeholder="Enter your last name"
                />
              </div>
            </div>
            
            {/* Bio */}
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">Bio</label>
              <Textarea
                value={formData.bio}
                onChange={(e) => setFormData(prev => ({ ...prev, bio: e.target.value }))}
                variant="bordered"
                isDisabled={!isEditing}
                placeholder="Tell us about yourself..."
                maxLength={280}
                className="min-h-24"
                classNames={{
                  input: "resize-none"
                }}
              />
              <div className="flex justify-between items-center">
                <p className="text-xs text-default-400">Share a bit about yourself with the community</p>
                <span className={`text-xs ${formData.bio.length > 250 ? 'text-warning' : 'text-default-400'}`}>
                  {formData.bio.length}/280
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        {isEditing && (
          <div className="flex gap-3 pt-6 border-t border-divider">
            <Button
              color="primary"
              onPress={handleSave}
              isLoading={isSaving}
              isDisabled={isSaving}
              className="font-medium px-8"
            >
              {isSaving ? 'Saving...' : 'Save Changes'}
            </Button>
            <Button
              variant="light"
              onPress={handleCancel}
              isDisabled={isSaving}
              className="font-medium px-8"
            >
              Cancel
            </Button>
          </div>
        )}
      </CardBody>
    </Card>
  )
}
