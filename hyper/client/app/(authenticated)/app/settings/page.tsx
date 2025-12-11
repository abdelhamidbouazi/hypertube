"use client";

import React, { useEffect, useState } from "react";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Avatar } from "@heroui/avatar";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import { Select, SelectItem } from "@heroui/select";
import { Divider } from "@heroui/divider";
import { addToast } from "@heroui/toast";
import { useMe, updateUser, uploadAvatar } from "@/lib/hooks";
import { getErrorMessage } from "@/lib/error-utils";
import { Eye, EyeOff, Save, User, Lock, Camera, Edit2 } from "lucide-react";

export default function SettingsPage() {
  const { user, isLoading: isUserLoading, refetch } = useMe();
  const [isLoading, setIsLoading] = useState(false);

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [language, setLanguage] = useState("en");

  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isPasswordVisible, setIsPasswordVisible] = useState(false);

  useEffect(() => {
    if (user) {
      setFirstName(user.firstname || "");
      setLastName(user.lastname || "");
      setUsername(user.username || "");
      setEmail(user.email || "");
      setLanguage(user.preferred_language || "en");
    }
  }, [user]);

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      await updateUser({
        firstname: firstName,
        lastname: lastName,
        username: username,
        email: email,
        preferred_language: language,
      });

      addToast({
        title: "Profile updated",
        description: "Your profile information has been saved.",
        severity: "success",
        timeout: 3000,
      });

      refetch();
    } catch (error) {
      addToast({
        title: "Update failed",
        description: getErrorMessage(error),
        severity: "danger",
        timeout: 4000,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handlePasswordUpdate = async (e: React.FormEvent) => {
    e.preventDefault();

    if (password !== confirmPassword) {
      addToast({
        title: "Passwords do not match",
        description: "Please ensure both password fields match.",
        severity: "warning",
        timeout: 3000,
      });
      return;
    }

    if (password.length < 8) {
      addToast({
        title: "Password too short",
        description: "Password must be at least 8 characters long.",
        severity: "warning",
        timeout: 3000,
      });
      return;
    }

    setIsLoading(true);

    try {
      await updateUser({
        password: password,
      });

      addToast({
        title: "Password updated",
        description: "Your password has been changed successfully.",
        severity: "success",
        timeout: 3000,
      });

      setPassword("");
      setConfirmPassword("");
    } catch (error) {
      addToast({
        title: "Update failed",
        description: getErrorMessage(error),
        severity: "danger",
        timeout: 4000,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsLoading(true);
    try {
      await uploadAvatar(file);
      addToast({
        title: "Avatar updated",
        description: "Your profile picture has been updated successfully.",
        severity: "success",
        timeout: 3000,
      });
      refetch();
    } catch (error) {
      addToast({
        title: "Upload failed",
        description: getErrorMessage(error),
        severity: "danger",
        timeout: 4000,
      });
    } finally {
      setIsLoading(false);
    }
  };

  if (isUserLoading) {
    return <div className="p-8 text-center">Loading settings...</div>;
  }

  return (
    <div className="container mx-auto px-4 pb-8 max-w-7xl">
      <h1 className="text-3xl font-bold mb-8 text-foreground">Settings</h1>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Profile Settings */}
        <Card className="p-4 h-full">
          <CardHeader className="flex gap-3">
            <User size={24} className="text-primary" />
            <div className="flex flex-col items-start">
              <p className="text-md font-bold">Profile Information</p>
              <p className="text-small text-default-500">
                Update your personal details
              </p>
            </div>
          </CardHeader>
          <Divider />
          <CardBody>
            <form onSubmit={handleProfileUpdate} className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="First Name"
                  placeholder="Enter your first name"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  variant="bordered"
                />
                <Input
                  label="Last Name"
                  placeholder="Enter your last name"
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                  variant="bordered"
                />
              </div>

              <Input
                label="Username"
                placeholder="Enter your username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                variant="bordered"
              />

              <Input
                label="Email"
                type="email"
                placeholder="Enter your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                variant="bordered"
              />

              <Select
                label="Preferred Language"
                selectedKeys={[language]}
                onChange={(e) => setLanguage(e.target.value)}
                variant="bordered"
              >
                <SelectItem key="en">English</SelectItem>
                <SelectItem key="fr">French</SelectItem>
                <SelectItem key="es">Spanish</SelectItem>
              </Select>

              <div className="flex justify-end">
                <Button
                  color="primary"
                  type="submit"
                  isLoading={isLoading}
                  startContent={<Save size={18} />}
                >
                  Save Profile
                </Button>
              </div>
            </form>
          </CardBody>
        </Card>

        {/* Password Settings */}
        <Card className="p-4 h-full">
          <CardHeader className="flex gap-3">
            <User size={24} className="text-primary" />
            <div className="flex flex-col items-start">
              <p className="text-md font-bold">Change Avatar</p>
              <p className="text-small text-default-500">
                Upload a new avatar image.
              </p>
            </div>
          </CardHeader>
          <Divider />
          <CardBody>
            <div className="flex flex-col items-center gap-4 mb-6">
              <div className="relative group cursor-pointer">
                <Avatar
                  src={user?.avatar}
                  className="w-24 h-24 text-large"
                  isBordered
                  color="primary"
                />
                <label
                  htmlFor="avatar-upload"
                  className="absolute inset-0 flex items-center justify-center bg-black/50 rounded-full opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
                >
                  <Edit2 className="text-white w-6 h-6" />
                </label>
                <input
                  id="avatar-upload"
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={handleAvatarUpload}
                  disabled={isLoading}
                />
              </div>
              <p className="text-small text-default-500">
                Click to change avatar
              </p>
            </div>
            <CardHeader className="flex gap-3">
              <Lock size={24} className="text-primary" />
              <div className="flex flex-col items-start">
                <p className="text-md font-bold">Change Password</p>
                <p className="text-small text-default-500">
                  Ensure your account is using a long password.
                </p>
              </div>
            </CardHeader>
            <Divider className="mb-6" />
            <form
              onSubmit={handlePasswordUpdate}
              className="space-y-6 h-full flex flex-col justify-between"
            >
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="New Password"
                  placeholder="Enter new password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  variant="bordered"
                  type={isPasswordVisible ? "text" : "password"}
                  endContent={
                    <button
                      className="focus:outline-none"
                      type="button"
                      onClick={() => setIsPasswordVisible(!isPasswordVisible)}
                    >
                      {isPasswordVisible ? (
                        <EyeOff className="text-2xl text-default-400 pointer-events-none" />
                      ) : (
                        <Eye className="text-2xl text-default-400 pointer-events-none" />
                      )}
                    </button>
                  }
                />
                <Input
                  label="Confirm Password"
                  placeholder="Confirm new password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  variant="bordered"
                  type={isPasswordVisible ? "text" : "password"}
                />
              </div>

              <div className="flex justify-end">
                <Button
                  color="primary"
                  type="submit"
                  isLoading={isLoading}
                  isDisabled={!password}
                  startContent={<Save size={18} />}
                >
                  Update Password
                </Button>
              </div>
            </form>
          </CardBody>
        </Card>
      </div>
    </div>
  );
}
