import { Component, ElementRef, ViewChild, AfterViewInit, OnDestroy, signal } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  standalone: true,
  selector: 'app-discover',
  imports: [CommonModule],
  templateUrl: './discover.page.html',
})
export class DiscoverPage implements AfterViewInit, OnDestroy {
  @ViewChild('infiniteScrollAnchor', { static: true }) anchorRef!: ElementRef<HTMLDivElement>;

  protected readonly movies = signal<DiscoverMovie[]>([]);
  protected readonly isLoading = signal<boolean>(false);
  protected readonly isEnd = signal<boolean>(false);

  private currentPage = 0;
  private readonly pageSize = 30;
  private observer: IntersectionObserver | null = null;

  constructor() {
    /** Kick off the first page load. */
    this.loadMore();
  }

  /** Attach an intersection observer to the sentinel to drive infinite scrolling. */
  ngAfterViewInit(): void {
    this.observer = new IntersectionObserver(
      entries => {
        const entry = entries[0];
        if (entry.isIntersecting) {
          this.loadMore();
        }
      },
      {
        root: null,
        rootMargin: '0px 0px 400px 0px',
        threshold: 0.01,
      }
    );

    if (this.anchorRef?.nativeElement) {
      this.observer.observe(this.anchorRef.nativeElement);
    }
  }

  /** Clean up the observer on destroy. */
  ngOnDestroy(): void {
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
  }

  /** Load the next page of mock movies with a short artificial delay. */
  protected loadMore(): void {
    if (this.isLoading() || this.isEnd()) return;
    this.isLoading.set(true);

    // Simulate async fetch latency
    setTimeout(() => {
      const next = this.generateMockMovies(this.currentPage, this.pageSize);
      if (next.length === 0) {
        this.isEnd.set(true);
      } else {
        this.movies.update(list => [...list, ...next]);
        this.currentPage += 1;
      }
      this.isLoading.set(false);
    }, 600);
  }

  /** Create a deterministic page of mock movies. Returns an empty array after 10 pages. */
  private generateMockMovies(page: number, count: number): DiscoverMovie[] {
    // Stop after 10 pages for demo
    if (page >= 10) return [];
    const baseIndex = page * count;
    const titles = [
      'Inception',
      'Interstellar',
      'The Matrix',
      'The Godfather',
      'The Dark Knight',
      'Parasite',
      'Whiplash',
      'Blade Runner 2049',
      'La La Land',
      'Arrival',
      'The Social Network',
      'Her',
      'Gladiator',
      'Fight Club',
      'Spirited Away',
      'City of God',
      'Se7en',
      'The Prestige',
      'Joker',
      'Mad Max: Fury Road',
    ];

    const items: DiscoverMovie[] = [];
    for (let i = 0; i < count; i++) {
      const idx = (baseIndex + i) % titles.length;
      const year = 1980 + ((baseIndex + i) % 44);
      const rating = Math.round((6 + Math.random() * 3.5) * 10) / 10; // range ~6.0–9.5
      const id = `${titles[idx].toLowerCase().replace(/[^a-z0-9]+/g, '-')}-${baseIndex + i}`;
      items.push({
        id,
        title: titles[idx],
        year,
        rating,
        imageUrl: `https://picsum.photos/seed/${encodeURIComponent(id)}/300/450`,
        description:
          'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore.',
      });
    }
    return items;
  }
}

interface DiscoverMovie {
  id: string;
  title: string;
  year: number;
  rating: number; // 0-10
  imageUrl: string;
  description: string;
}


