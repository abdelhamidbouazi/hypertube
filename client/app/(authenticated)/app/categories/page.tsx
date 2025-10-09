import Link from "next/link";

type Category = {
  name: string;
  slug: string;
  emoji: string;
  tone: string; // gradient tailwind classes
  blurb: string;
};

const CATEGORIES: Category[] = [
  { name: "Action", slug: "action", emoji: "💥", tone: "from-rose-500 to-orange-500", blurb: "Adrenaline, scale, and high-stakes set pieces." },
  { name: "Adventure", slug: "adventure", emoji: "🧭", tone: "from-emerald-500 to-lime-500", blurb: "Journeys across worlds, maps, and myths." },
  { name: "Comedy", slug: "comedy", emoji: "😂", tone: "from-yellow-500 to-amber-500", blurb: "Timing, wit, and laugh-out-loud beats." },
  { name: "Drama", slug: "drama", emoji: "🎭", tone: "from-violet-500 to-fuchsia-500", blurb: "Human stories with heart and conflict." },
  { name: "Sci-Fi", slug: "sci-fi", emoji: "🚀", tone: "from-cyan-500 to-sky-500", blurb: "Futures, frontiers, and thought experiments." },
  { name: "Horror", slug: "horror", emoji: "👻", tone: "from-red-500 to-rose-600", blurb: "Atmosphere, dread, and midnight chills." },
  { name: "Romance", slug: "romance", emoji: "💞", tone: "from-pink-500 to-rose-500", blurb: "Chemistry, longing, and soft landings." },
  { name: "Thriller", slug: "thriller", emoji: "🗝️", tone: "from-slate-600 to-slate-800", blurb: "Twists, turns, and razor-wire tension." },
  { name: "Animation", slug: "animation", emoji: "🎨", tone: "from-indigo-500 to-blue-500", blurb: "Craft, color, and imagination in motion." },
  { name: "Documentary", slug: "documentary", emoji: "🎥", tone: "from-stone-500 to-neutral-600", blurb: "True stories told with clarity." },
  { name: "Fantasy", slug: "fantasy", emoji: "🪄", tone: "from-purple-500 to-indigo-600", blurb: "Magic systems, quests, and wonder." },
  { name: "Crime", slug: "crime", emoji: "🕵️‍♂️", tone: "from-zinc-600 to-gray-700", blurb: "Mystery, motive, and moral gray." },
];

export default function CategoriesPage() {
  return (
    <article className="space-y-8 text-left">
      {/* Hero */}
      <header className="relative overflow-hidden rounded-3xl border border-gray-200 bg-white/60 p-6 shadow-sm backdrop-blur dark:border-gray-800 dark:bg-gray-900/60">
        <div className="absolute inset-0 -z-10 opacity-40 blur-3xl">
          <div className="absolute -top-24 right-0 h-48 w-48 rounded-full bg-gradient-to-br from-indigo-500 to-pink-500 dark:from-indigo-400 dark:to-pink-400" />
          <div className="absolute -bottom-24 left-0 h-48 w-48 rounded-full bg-gradient-to-br from-cyan-500 to-sky-500" />
        </div>
        <h1 className="text-3xl font-extrabold tracking-tight">
          <span className="bg-gradient-to-r from-indigo-600 to-pink-600 bg-clip-text text-transparent dark:from-indigo-400 dark:to-pink-400">
            Browse by Category
          </span>
        </h1>
        <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
          Curated genres with elegant navigation. Light &amp; dark ready.
        </p>
        <div className="mt-4 flex flex-wrap gap-2">
          <span className="rounded-full border border-gray-200 bg-gray-50 px-3 py-1 text-xs text-gray-700 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200">Shiny UI</span>
          <span className="rounded-full border border-gray-200 bg-gray-50 px-3 py-1 text-xs text-gray-700 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200">Accessible</span>
          <span className="rounded-full border border-gray-200 bg-gray-50 px-3 py-1 text-xs text-gray-700 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200">Tailwind-only</span>
        </div>
      </header>

      {/* Pro List (one-per-row to look premium inside max-w-lg) */}
      <ul className="space-y-4">
        {CATEGORIES.map((cat) => (
          <li key={cat.slug}>
            <Link
              href={`/app/categories/${encodeURIComponent(cat.slug)}`}
              className="group relative block overflow-hidden rounded-2xl border border-gray-200 bg-white/70 p-4 shadow-sm ring-1 ring-transparent backdrop-blur transition hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:border-gray-800 dark:bg-gray-900/60"
            >
              {/* Accent */}
              <div className="absolute inset-0 -z-10 opacity-0 blur-2xl transition-opacity duration-500 group-hover:opacity-40">
                <div className={`absolute inset-0 bg-gradient-to-br ${cat.tone}`} />
              </div>

              <div className="flex items-start gap-4">
                {/* Icon Badge */}
                <div
                  className={`shrink-0 rounded-2xl p-3 text-xl text-white shadow-md ring-1 ring-white/20 dark:ring-white/10 bg-gradient-to-br ${cat.tone} transition-transform duration-300 group-hover:scale-105`}
                  aria-hidden
                >
                  {cat.emoji}
                </div>

                {/* Content */}
                <div className="min-w-0">
                  <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                    {cat.name}
                  </h2>
                  <p className="mt-0.5 line-clamp-2 text-sm text-gray-600 dark:text-gray-400">
                    {cat.blurb}
                  </p>

                  {/* Meta */}
                  <div className="mt-3 flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                    <span className="rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5 dark:border-gray-800 dark:bg-gray-800">
                      Movies
                    </span>
                    <span className="rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5 dark:border-gray-800 dark:bg-gray-800">
                      Thumbnails
                    </span>
                    <span className="rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5 dark:border-gray-800 dark:bg-gray-800">
                      Sort &amp; Filter
                    </span>
                  </div>
                </div>

                {/* Arrow */}
                <div className="ml-auto flex h-9 w-9 items-center justify-center rounded-full border border-gray-200 text-gray-500 transition group-hover:translate-x-0.5 group-hover:text-gray-700 dark:border-gray-800 dark:text-gray-400">
                  <svg
                    className="h-4 w-4"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    viewBox="0 0 24 24"
                    aria-hidden
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                  </svg>
                </div>
              </div>
            </Link>
          </li>
        ))}
      </ul>

      <footer className="pt-2 text-center text-xs text-gray-500 dark:text-gray-400">
        Choose a category to view its movies.
      </footer>
    </article>
  );
}
