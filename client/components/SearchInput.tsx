"use client";

import React, { useRef, useEffect } from "react";
import { Input } from "@heroui/input";
import { Kbd } from "@heroui/kbd";

import { SearchResults } from "./SearchResults";
import { SearchIcon } from "./icons";

import { useSearch } from "@/hooks/useSearch";

interface SearchInputProps {
  className?: string;
  placeholder?: string;
  showKbd?: boolean;
  isMobile?: boolean;
}

export function SearchInput({
  className = "",
  placeholder = "Search movies...",
  showKbd = false,
  isMobile = false,
}: SearchInputProps) {
  const {
    query,
    setQuery,
    results,
    isLoading,
    error,
    isOpen,
    setIsOpen,
    clearSearch,
  } = useSearch();

  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Handle clicking outside to close results
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen, setIsOpen]);

  // Handle keyboard shortcuts
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      // Cmd+K or Ctrl+K to focus search
      if ((event.metaKey || event.ctrlKey) && event.key === "k") {
        event.preventDefault();
        inputRef.current?.focus();
      }

      // Escape to clear/close search
      if (event.key === "Escape") {
        if (isOpen || query) {
          event.preventDefault();
          clearSearch();
          inputRef.current?.blur();
        }
      }
    }

    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen, query, clearSearch]);

  const handleInputChange = (value: string) => {
    setQuery(value);
  };

  const handleInputFocus = () => {
    if (query.trim() && results.length > 0) {
      setIsOpen(true);
    }
  };

  const handleClose = () => {
    setIsOpen(false);
  };

  const inputHeight = isMobile ? "h-8" : "h-10";

  return (
    <div ref={containerRef} className={`relative ${className}`}>
      <Input
        ref={inputRef}
        aria-label="Search"
        classNames={{
          inputWrapper: `bg-default-100 ${inputHeight}`,
          input: "text-sm",
        }}
        endContent={
          showKbd ? (
            <Kbd className="hidden lg:inline-block" keys={["command"]}>
              K
            </Kbd>
          ) : undefined
        }
        labelPlacement="outside"
        placeholder={placeholder}
        startContent={
          <SearchIcon className="text-base text-default-400 pointer-events-none flex-shrink-0" />
        }
        type="search"
        value={query}
        onFocus={handleInputFocus}
        onValueChange={handleInputChange}
      />

      {isOpen && (
        <SearchResults
          error={error}
          isLoading={isLoading}
          query={query}
          results={results}
          onClose={handleClose}
        />
      )}
    </div>
  );
}
