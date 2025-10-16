"use client";

import { Suspense } from "react";

import SearchPageContent from "./SearchPageContent";

export default function SearchPage() {
  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8">
        <Suspense fallback={<div>Loading search results...</div>}>
          <SearchPageContent />
        </Suspense>
      </div>
    </div>
  );
}
