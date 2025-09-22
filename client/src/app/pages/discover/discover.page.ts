import { Component, ElementRef, ViewChild, AfterViewInit, OnDestroy, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { NzCarouselModule, NzCarouselComponent } from 'ng-zorro-antd/carousel';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzSliderModule } from 'ng-zorro-antd/slider';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzTagModule } from 'ng-zorro-antd/tag';
import { NzRateModule } from 'ng-zorro-antd/rate';
import { NzSpinModule } from 'ng-zorro-antd/spin';

@Component({
  standalone: true,
  selector: 'app-discover',
  imports: [
    CommonModule,
    FormsModule,
    NzCarouselModule,
    NzGridModule,
    NzCardModule,
    NzIconModule,
    NzInputModule,
    NzSelectModule,
    NzSliderModule,
    NzButtonModule,
    NzTagModule,
    NzRateModule,
    NzSpinModule,
  ],
  templateUrl: './discover.page.html',
})
export class DiscoverPage implements AfterViewInit, OnDestroy {
  protected readonly Math = Math;
  @ViewChild('infiniteScrollAnchor', { static: true }) anchorRef!: ElementRef<HTMLDivElement>;
  @ViewChild(NzCarouselComponent) heroCarousel?: NzCarouselComponent;

  protected readonly movies = signal<DiscoverMovie[]>([]);
  protected readonly isLoading = signal<boolean>(false);
  protected readonly isEnd = signal<boolean>(false);

  // Active slide index for hero carousel
  protected readonly heroActiveIndex = signal<number>(0);

  // Filters
  protected readonly searchQuery = signal<string>('');
  protected readonly selectedGenres = signal<string[]>([]);
  protected readonly yearRange = signal<[number, number]>([1990, 2024]);
  protected readonly minRating = signal<number>(0);
  protected readonly sortBy = signal<'rating-desc' | 'year-desc' | 'title-asc'>('rating-desc');

  protected readonly allGenres = [
    'Action',
    'Adventure',
    'Animation',
    'Comedy',
    'Crime',
    'Drama',
    'Fantasy',
    'History',
    'Horror',
    'Mystery',
    'Romance',
    'Sci-Fi',
    'Thriller',
  ];

  protected readonly minYear = 1970;
  protected readonly maxYear = new Date().getFullYear();

  /** Movies after applying filters and sorting */
  protected readonly filteredMovies = computed(() => {
    const q = this.searchQuery().trim().toLowerCase();
    const [from, to] = this.yearRange();
    const genres = this.selectedGenres();
    const minR = this.minRating();

    let list = this.movies().filter(m => {
      const matchesQuery = q === '' || m.title.toLowerCase().includes(q);
      const matchesYear = m.year >= from && m.year <= to;
      const matchesRating = m.rating >= minR;
      const matchesGenre = genres.length === 0 || genres.some(g => m.genres.includes(g));
      return matchesQuery && matchesYear && matchesRating && matchesGenre;
    });

    switch (this.sortBy()) {
      case 'year-desc':
        list = [...list].sort((a, b) => b.year - a.year);
        break;
      case 'title-asc':
        list = [...list].sort((a, b) => a.title.localeCompare(b.title));
        break;
      default:
        list = [...list].sort((a, b) => b.rating - a.rating);
    }
    return list;
  });

  protected readonly heroSlides = computed(() => {
    return [...this.movies()]
      .sort((a, b) => b.rating - a.rating)
      .slice(0, 8);
  });

  /** Carousel controls */
  protected nextHero(): void {
    this.heroCarousel?.next();
  }

  protected prevHero(): void {
    this.heroCarousel?.pre();
  }

  protected goToHero(index: number): void {
    if (index < 0) return;
    this.heroCarousel?.goTo(index);
    this.heroActiveIndex.set(index);
  }

  protected onHeroBeforeChange(evt: { from: number; to: number }): void {
    this.heroActiveIndex.set(evt.to);
  }

  private currentPage = 0;
  private readonly pageSize = 30;
  private observer: IntersectionObserver | null = null;

  constructor() {
    this.loadMore();
  }

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


  private generateMockMovies(page: number, count: number): DiscoverMovie[] {
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
    const genres = this.allGenres;

    const items: DiscoverMovie[] = [];
    for (let i = 0; i < count; i++) {
      const idx = (baseIndex + i) % titles.length;
      const year = 1980 + ((baseIndex + i) % 44);
      const rating = Math.round((6 + Math.random() * 3.5) * 10) / 10; 
      const id = `${titles[idx].toLowerCase().replace(/[^a-z0-9]+/g, '-')}-${baseIndex + i}`;
      const movieGenres = [genres[(baseIndex + i) % genres.length], genres[(baseIndex + i + 3) % genres.length]];
      items.push({
        id,
        title: titles[idx],
        year,
        rating,
        imageUrl: `https://picsum.photos/seed/${encodeURIComponent(id)}/300/450`,
        description:
          'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore.',
        genres: movieGenres,
      });
    }
    return items;
  }

  protected clearFilters(): void {
    this.searchQuery.set('');
    this.selectedGenres.set([]);
    this.yearRange.set([1990, this.maxYear]);
    this.minRating.set(0);
    this.sortBy.set('rating-desc');
  }

  /** Map 0–5 star input (with halves) to 0–10 rating scale and update filter */
  protected onRateChange(stars: number): void {
    const mapped = Math.round(stars * 2 * 10) / 10;
    this.minRating.set(mapped);
  }

  /** Remove one genre from the selected list (used by closable tags) */
  protected removeGenre(genre: string, event?: MouseEvent): void {
    if (event) event.preventDefault();
    this.selectedGenres.set(this.selectedGenres().filter(g => g !== genre));
  }
}

interface DiscoverMovie {
  id: string;
  title: string;
  year: number;
  rating: number; // 0-10
  imageUrl: string;
  description: string;
  genres: string[];
}


