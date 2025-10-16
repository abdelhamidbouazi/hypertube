"use client";
import {
  Navbar as HeroUINavbar,
  NavbarContent,
  NavbarMenu,
  NavbarMenuToggle,
  NavbarBrand,
  NavbarItem,
  NavbarMenuItem,
} from "@heroui/navbar";
import { Link } from "@heroui/link";
import { Avatar } from "@heroui/avatar";
import {
  Dropdown,
  DropdownTrigger,
  DropdownMenu,
  DropdownItem,
} from "@heroui/dropdown";
import { link as linkStyles } from "@heroui/theme";
import NextLink from "next/link";
import clsx from "clsx";
import { User, Settings, LogOut, Bookmark, History } from "lucide-react";

import { siteConfig } from "@/config/site";
import { ThemeSwitch } from "@/components/theme-switch";
import { SearchInput } from "@/components/SearchInput";
import { logoutUser } from "@/lib/hooks";
import { useAuth } from "@/lib/hooks";
import { TwitterIcon, GithubIcon, DiscordIcon, Logo } from "@/components/icons";

export const Navbar = () => {
  const { user } = useAuth();

  const desktopSearchInput = (
    <SearchInput
      className="w-64 md:w-72 lg:w-80"
      placeholder="Search movies..."
      showKbd={true}
    />
  );

  const mobileSearchInput = (
    <SearchInput
      className="w-full max-w-xs"
      isMobile={true}
      placeholder="Search movies..."
    />
  );

  const isAuthenticated = !!user;
  const displayName = isAuthenticated
    ? [user?.firstname, user?.lastname].filter(Boolean).join(" ") ||
      (user?.email ? user.email.split("@")[0] : "User")
    : "User";
  const avatarSrc =
    user?.avatar ||
    (user?.email
      ? `https://i.pravatar.cc/150?u=${encodeURIComponent(user.email)}`
      : undefined);
  const signedInLabel = user
    ? [user?.firstname, user?.lastname].filter(Boolean).join(" ") ||
      user?.email ||
      ""
    : "";

  return (
    <HeroUINavbar isBordered maxWidth="xl" position="sticky">
      <NavbarContent className="basis-1/3" justify="start">
        <NavbarBrand as="li" className="gap-2 max-w-fit">
          <NextLink
            className="flex justify-start items-center gap-2"
            href={isAuthenticated ? "/app/discover" : "/"}
          >
            <Logo />
            <p className="text-inherit text-2xl lg:text-3xl font-light hidden sm:block">
              CINÉTHOS
            </p>
          </NextLink>
        </NavbarBrand>

        <ul className="hidden lg:flex gap-4 justify-start ml-4">
          {siteConfig.navItems.map((item) => (
            <NavbarItem key={item.href}>
              <NextLink
                className={clsx(
                  linkStyles({ color: "foreground" }),
                  "data-[active=true]:text-primary data-[active=true]:font-medium"
                )}
                color="foreground"
                href={item.href}
              >
                {item.label}
              </NextLink>
            </NavbarItem>
          ))}
        </ul>
      </NavbarContent>

      <NavbarContent className="hidden md:flex basis-1/3" justify="center">
        <NavbarItem className="w-full max-w-sm">
          {isAuthenticated && desktopSearchInput}
        </NavbarItem>
      </NavbarContent>

      <NavbarContent className="hidden md:flex basis-1/3" justify="end">
        <NavbarItem className="hidden lg:flex gap-2">
          <Link isExternal aria-label="X" href={siteConfig.links.X}>
            <TwitterIcon className="text-default-500 hover:text-default-700 transition-colors" />
          </Link>
          <Link
            isExternal
            aria-label="Portfolio"
            href={siteConfig.links.porfolio}
          >
            <DiscordIcon className="text-default-500 hover:text-default-700 transition-colors" />
          </Link>
          <Link isExternal aria-label="Github" href={siteConfig.links.github}>
            <GithubIcon className="text-default-500 hover:text-default-700 transition-colors" />
          </Link>
          <ThemeSwitch />
        </NavbarItem>
        <NavbarItem>
          <Dropdown placement="bottom-end">
            <DropdownTrigger>
              <Avatar
                isBordered
                as="button"
                className="transition-transform hover:scale-105"
                color="primary"
                imgProps={{
                  crossOrigin: "anonymous",
                  referrerPolicy: "no-referrer",
                }}
                name={displayName}
                size="sm"
                src={avatarSrc}
              />
            </DropdownTrigger>
            <DropdownMenu aria-label="Profile Actions" variant="flat">
              <DropdownItem
                key="user-info"
                className="h-14 gap-2"
                textValue="User info"
              >
                <p className="font-semibold">Signed in as</p>
                <p className="font-semibold text-primary">{signedInLabel}</p>
              </DropdownItem>
              <DropdownItem
                key="profile"
                startContent={<User className="h-4 w-4" />}
              >
                My Profile
              </DropdownItem>
              <DropdownItem
                key="bookmarks"
                startContent={<Bookmark className="h-4 w-4" />}
              >
                My Bookmarks
              </DropdownItem>
              <DropdownItem
                key="history"
                startContent={<History className="h-4 w-4" />}
              >
                Watch History
              </DropdownItem>
              <DropdownItem
                key="settings"
                startContent={<Settings className="h-4 w-4" />}
              >
                Settings
              </DropdownItem>
              <DropdownItem
                key="logout"
                color="danger"
                startContent={<LogOut className="h-4 w-4" />}
                onPress={() => logoutUser()}
              >
                Log Out
              </DropdownItem>
            </DropdownMenu>
          </Dropdown>
        </NavbarItem>
      </NavbarContent>

      <NavbarContent className="md:hidden" justify="end">
        <NavbarItem>
          <ThemeSwitch />
        </NavbarItem>
        <NavbarItem>
          <Dropdown placement="bottom-end">
            <DropdownTrigger>
              <Avatar
                isBordered
                as="button"
                className="transition-transform"
                color="primary"
                imgProps={{
                  crossOrigin: "anonymous",
                  referrerPolicy: "no-referrer",
                }}
                name={displayName}
                size="sm"
                src={avatarSrc}
              />
            </DropdownTrigger>
            <DropdownMenu aria-label="Profile Actions" variant="flat">
              <DropdownItem
                key="user-info"
                className="h-14 gap-2"
                textValue="User info"
              >
                <p className="font-semibold">Signed in as</p>
                <p className="font-semibold text-primary">{signedInLabel}</p>
              </DropdownItem>
              <DropdownItem
                key="profile"
                startContent={<User className="h-4 w-4" />}
              >
                My Profile
              </DropdownItem>
              <DropdownItem
                key="bookmarks"
                startContent={<Bookmark className="h-4 w-4" />}
              >
                My Bookmarks
              </DropdownItem>
              <DropdownItem
                key="history"
                startContent={<History className="h-4 w-4" />}
              >
                Watch History
              </DropdownItem>
              <DropdownItem
                key="settings"
                startContent={<Settings className="h-4 w-4" />}
              >
                Settings
              </DropdownItem>
              <DropdownItem
                key="logout"
                color="danger"
                startContent={<LogOut className="h-4 w-4" />}
                onPress={() => logoutUser()}
              >
                Log Out
              </DropdownItem>
            </DropdownMenu>
          </Dropdown>
        </NavbarItem>
        <NavbarMenuToggle />
      </NavbarContent>

      <NavbarMenu>
        <div className="mx-4 mt-4 flex flex-col gap-4">
          {isAuthenticated && (
            <div className="pb-2 border-b border-default-200">
              {mobileSearchInput}
            </div>
          )}

          <div className="flex flex-col gap-2">
            {siteConfig.navItems.map((item) => (
              <NavbarMenuItem key={item.href}>
                <NextLink
                  className="w-full text-foreground text-lg hover:text-primary transition-colors"
                  href={item.href}
                >
                  {item.label}
                </NextLink>
              </NavbarMenuItem>
            ))}
          </div>

          {/* <div className="pt-2 border-t border-default-200">
            {siteConfig.navMenuItems.map((item, index) => (
              <NavbarMenuItem key={`${item.label}-${index}`}>
                <Link
                  className="w-full"
                  color={
                    index === siteConfig.navMenuItems.length - 1
                      ? "danger"
                      : "foreground"
                  }
                  href={item.href}
                  size="lg"
                >
                  {item.label}
                </Link>
              </NavbarMenuItem>
            ))}
          </div> */}

          <div className="pt-2 border-t border-default-200">
            <p className="text-sm text-default-500 mb-3">Connect with us</p>
            <div className="flex gap-4">
              <Link isExternal aria-label="X" href={siteConfig.links.X}>
                <TwitterIcon className="text-default-500 hover:text-default-700 transition-colors w-5 h-5" />
              </Link>
              <Link
                isExternal
                aria-label="Portfolio"
                href={siteConfig.links.porfolio}
              >
                <DiscordIcon className="text-default-500 hover:text-default-700 transition-colors w-5 h-5" />
              </Link>
              <Link
                isExternal
                aria-label="Github"
                href={siteConfig.links.github}
              >
                <GithubIcon className="text-default-500 hover:text-default-700 transition-colors w-5 h-5" />
              </Link>
            </div>
          </div>
        </div>
      </NavbarMenu>
    </HeroUINavbar>
  );
};
