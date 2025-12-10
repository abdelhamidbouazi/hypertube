"use client";

import { useEffect } from "react";
import { Card, CardBody } from "@heroui/card";
import { Button } from "@heroui/button";
import { Alert } from "@heroui/alert";
import { Chip } from "@heroui/chip";
import { addToast } from "@heroui/toast";
import { getErrorMessage } from "@/lib/error-utils";
import {
  AlertTriangle,
  RefreshCw,
  Home,
  ArrowLeft,
  Bug,
  Wifi,
  Server,
} from "lucide-react";
import Link from "next/link";

export default function Error({
  error,
  reset,
}: {
  error: Error;
  reset: () => void;
}) {
  useEffect(() => {
    addToast({
      title: "An error occurred",
      description: getErrorMessage(error),
      severity: "danger",
      timeout: 5000,
    });
  }, [error]);

  const getErrorType = (error: Error) => {
    const message = error.message.toLowerCase();
    if (message.includes("network") || message.includes("fetch"))
      return "network";
    if (message.includes("server") || message.includes("500")) return "server";
    if (message.includes("not found") || message.includes("404"))
      return "notfound";
    return "general";
  };

  const errorType = getErrorType(error);

  const errorConfig = {
    network: {
      icon: <Wifi size={48} className="text-warning" />,
      title: "Connection Lost",
      description:
        "It looks like you've lost your internet connection. Please check your network and try again.",
      color: "warning" as const,
      suggestions: [
        "Check your internet connection",
        "Try refreshing the page",
        "Switch to a different network",
      ],
    },
    server: {
      icon: <Server size={48} className="text-danger" />,
      title: "Server Error",
      description:
        "Our servers are experiencing some issues. We're working to fix this as quickly as possible.",
      color: "danger" as const,
      suggestions: [
        "Try again in a few minutes",
        "Check our status page",
        "Contact support if the issue persists",
      ],
    },
    notfound: {
      icon: <AlertTriangle size={48} className="text-warning" />,
      title: "Page Not Found",
      description:
        "The page you're looking for doesn't exist or has been moved.",
      color: "warning" as const,
      suggestions: [
        "Check the URL for typos",
        "Go back to the previous page",
        "Visit our homepage",
      ],
    },
    general: {
      icon: <Bug size={48} className="text-danger" />,
      title: "Something Went Wrong",
      description:
        "An unexpected error occurred. Don't worry, we've been notified and are looking into it.",
      color: "danger" as const,
      suggestions: [
        "Try refreshing the page",
        "Clear your browser cache",
        "Contact support if the issue persists",
      ],
    },
  };

  const config = errorConfig[errorType];

  return (
    <div className="min-h-screen bg-gradient-to-br from-background to-content2 flex items-center justify-center p-4">
      <div className="w-full max-w-2xl">
        {/* Background Effects */}
        <div className="relative overflow-hidden rounded-3xl border border-default-200 bg-content1/70 p-8 shadow-sm backdrop-blur-md dark:border-default-100">
          <div className="pointer-events-none absolute inset-0 -z-10 opacity-40 blur-3xl">
            <div className="absolute -top-24 right-0 h-48 w-48 rounded-full bg-gradient-to-br from-red-500 to-orange-500 dark:from-red-400 dark:to-orange-400" />
            <div className="absolute -bottom-24 left-0 h-48 w-48 rounded-full bg-gradient-to-br from-purple-500 to-pink-500" />
          </div>

          {/* Error Content */}
          <div className="text-center space-y-6">
            {/* Error Icon */}
            <div className="flex justify-center">{config.icon}</div>

            {/* Error Title */}
            <div>
              <h1 className="text-4xl font-extrabold tracking-tight mb-2">
                <span className="bg-gradient-to-r from-red-600 to-orange-600 bg-clip-text text-transparent dark:from-red-400 dark:to-orange-400">
                  {config.title}
                </span>
              </h1>
              <p className="text-lg text-foreground-600 max-w-md mx-auto">
                {config.description}
              </p>
            </div>

            {/* Error Details */}
            <Card className="bg-content2/50 backdrop-blur-sm">
              <CardBody className="p-4">
                <div className="flex items-center justify-center gap-2 mb-3">
                  <Bug size={16} className="text-foreground-500" />
                  <span className="text-sm font-medium text-foreground-500">
                    Error Details
                  </span>
                </div>
                <code className="text-xs text-foreground-600 bg-content1 p-2 rounded border block text-left break-all">
                  {error.message || "Unknown error occurred"}
                </code>
              </CardBody>
            </Card>

            {/* Suggestions */}
            <div className="space-y-3">
              <h3 className="text-lg font-semibold text-foreground-700">
                What you can try:
              </h3>
              <div className="flex flex-wrap justify-center gap-2">
                {config.suggestions.map((suggestion, index) => (
                  <Chip
                    key={index}
                    variant="flat"
                    color={config.color}
                    size="sm"
                    className="text-xs"
                  >
                    {suggestion}
                  </Chip>
                ))}
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-3 justify-center pt-4">
              <Button
                variant="flat"
                color="success"
                size="lg"
                startContent={<ArrowLeft size={20} />}
                onPress={() => window.history.back()}
                className="min-w-[140px]"
              >
                Go Back
              </Button>

              <Button
                color="primary"
                variant="flat"
                size="lg"
                startContent={<RefreshCw size={20} />}
                onPress={reset}
                className="min-w-[140px]"
              >
                Try Again
              </Button>

              <Button
                as={Link}
                href="/app/discover"
                color="secondary"
                variant="flat"
                size="lg"
                startContent={<Home size={20} />}
                className="min-w-[140px]"
              >
                Home
              </Button>
            </div>

            {/* Additional Help */}
            <div className="pt-6 border-t border-default-200">
              <p className="text-sm text-foreground-500 mb-3">
                Still having trouble? We&apos;re here to help.
              </p>
              <div className="flex justify-center gap-4">
                <Button
                  as={Link}
                  href="/app/settings"
                  variant="light"
                  size="sm"
                  className="text-xs"
                >
                  Contact Support (stamim)
                </Button>
                <Button
                  as={Link}
                  href="/app/about"
                  variant="light"
                  size="sm"
                  className="text-xs"
                >
                  Help Center - (amessah)
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
