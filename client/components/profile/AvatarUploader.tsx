'use client'

import { useState, useRef } from 'react'
import { Avatar } from '@heroui/avatar'
import { Button } from '@heroui/button'
import { Card, CardBody } from '@heroui/card'
import { toast } from 'react-hot-toast'

interface AvatarUploaderProps {
  src?: string
  name: string
  onAvatarChange: (file: File | null) => void
  onRemove: () => void
}

export default function AvatarUploader({ src, name, onAvatarChange, onRemove }: AvatarUploaderProps) {
  const [preview, setPreview] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    // Validate file type
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp']
    if (!allowedTypes.includes(file.type)) {
      toast.error('Please select a valid image file (.jpg, .png, .webp)')
      return
    }

    // Validate file size (2MB max)
    const maxSize = 2 * 1024 * 1024 // 2MB in bytes
    if (file.size > maxSize) {
      toast.error('File size must be less than 2MB')
      return
    }

    // Create preview
    const reader = new FileReader()
    reader.onload = (e) => {
      const result = e.target?.result as string
      setPreview(result)
      onAvatarChange(file)
    }
    reader.readAsDataURL(file)
  }

  const handleRemove = () => {
    setPreview(null)
    onAvatarChange(null)
    onRemove()
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const displaySrc = preview || src

  return (
    <div className="flex flex-col items-center space-y-4">
      <div className="relative">
        <Avatar
          size="lg"
          className="w-24 h-24 ring-4 ring-primary/10 shadow-xl"
          src={displaySrc}
          name={name}
        />
        {displaySrc && (
          <div className="absolute -top-2 -right-2">
            <div className="w-6 h-6 bg-success rounded-full flex items-center justify-center">
              <span className="text-white text-xs">✓</span>
            </div>
          </div>
        )}
      </div>
      
      <div className="flex flex-col items-center space-y-3">
        <div className="flex gap-2">
          <Button
            size="sm"
            color="primary"
            variant="flat"
            onPress={() => fileInputRef.current?.click()}
            className="font-medium"
          >
            {displaySrc ? 'Change' : 'Upload'}
          </Button>
          {displaySrc && (
            <Button
              size="sm"
              variant="light"
              color="danger"
              onPress={handleRemove}
              className="font-medium"
            >
              Remove
            </Button>
          )}
        </div>

        <div className="text-center">
          <p className="text-xs text-default-500 font-medium">Profile Photo</p>
          <p className="text-xs text-default-400 mt-1">
            JPG, PNG, WEBP up to 2MB
          </p>
        </div>
      </div>

      <input
        ref={fileInputRef}
        type="file"
        accept=".jpg,.jpeg,.png,.webp"
        onChange={handleFileSelect}
        className="hidden"
      />
    </div>
  )
}
