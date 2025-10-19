'use client'
import {
  Navbar as HeroUINavbar,
  NavbarContent,
  NavbarMenu,
  NavbarMenuToggle,
  NavbarBrand,
  NavbarItem,
  NavbarMenuItem,
} from "@heroui/navbar";
import { Button } from "@heroui/button";
import { Kbd } from "@heroui/kbd";
import { Link } from "@heroui/link";
import { Input } from "@heroui/input";
import { Avatar } from "@heroui/avatar";
import { Dropdown, DropdownTrigger, DropdownMenu, DropdownItem } from "@heroui/dropdown";
import { link as linkStyles } from "@heroui/theme";
import NextLink from "next/link";
import clsx from "clsx";

import { siteConfig } from "@/config/site";
import { ThemeSwitch } from "@/components/theme-switch";
import { logoutUser } from "@/lib/hooks";
import {
  TwitterIcon,
  GithubIcon,
  DiscordIcon,
  SearchIcon,
  Logo,
  LoginIcon,
} from "@/components/icons";
import { useState } from "react";
import { User, Settings, LogOut, Bookmark, History, Menu } from "lucide-react";

export const Navbar = ({ 
  onSidebarToggle, 
  onSidebarCollapseToggle, 
  isSidebarCollapsed 
}: {
  onSidebarToggle?: () => void
  onSidebarCollapseToggle?: () => void
  isSidebarCollapsed?: boolean
}) => {
  // search input component
  const searchInput = (
    <Input
      variant="flat"
      color="primary"
      aria-label="Search"
      classNames={{
        inputWrapper: "bg-content1/50 backdrop-blur-sm border-primary/20 hover:border-primary/40 focus-within:border-primary shadow-sm",
        input: "text-foreground placeholder:text-default-500"
      }}
      endContent={
        <Kbd className="hidden lg:inline-block bg-primary/10 text-primary border-primary/20" keys={["command"]}>
          K
        </Kbd>
      }
      labelPlacement="outside"
      placeholder="Search movies..."
      startContent={
        <SearchIcon className="text-base text-primary/70 pointer-events-none flex-shrink-0" />
      }
      type="search"
    />
  );
  const [isAuthenticated, setIsAuthenticated] = useState(true)

  return (
    <HeroUINavbar position="sticky">
      {/* left side - empty for balance */}
      <NavbarContent justify="start">
        <NavbarItem>
        </NavbarItem>
      </NavbarContent>

      {/* center - search input */}
      <NavbarContent justify="center">
        <NavbarItem className="hidden md:flex">
          {searchInput}
        </NavbarItem>
      </NavbarContent>

      {/* right side - empty for balance */}
      <NavbarContent justify="end">
        <NavbarItem>
        </NavbarItem>
      </NavbarContent>

      {/* mobile menu toggle */}
      <NavbarContent className="sm:hidden" justify="end">
        <NavbarMenuToggle />
      </NavbarContent>

      {/* mobile menu content */}
      <NavbarMenu>
        {searchInput}
        <div className="mx-4 mt-2 flex flex-col gap-2">
          {siteConfig.navMenuItems.map((item, index) => (
            <NavbarMenuItem key={`${item}-${index}`}>
              <Link
                color={
                  index === 2
                    ? "primary"
                    : index === siteConfig.navMenuItems.length - 1
                      ? "danger"
                      : "foreground"
                }
                href="#"
                size="lg"
              >
                {item.label}
              </Link>
            </NavbarMenuItem>
          ))}
        </div>
      </NavbarMenu>
    </HeroUINavbar>
  );
};
