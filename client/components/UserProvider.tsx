"use client";

import React from "react";

import { useAuth } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

export function UserProvider({ children }: { children: React.ReactNode }) {
  const { user: apiUser, isLoading } = useAuth();
  const { setUser, user: storeUser } = useAuthStore();

  React.useEffect(() => {
    if (apiUser && !isLoading && (!storeUser || storeUser.id !== apiUser.id)) {
      setUser(apiUser);
    }
  }, [apiUser, isLoading, storeUser, setUser]);

  return <>{children}</>;
}
