export default function AboutPage() {
  return (
    <article className="space-y-8 text-left">
      {/* Title */}
      <header className="space-y-2">
        <h1 className="text-3xl font-extrabold tracking-tight">
          <span className="bg-gradient-to-r from-indigo-600 to-pink-600 bg-clip-text text-transparent dark:from-indigo-400 dark:to-pink-400">
            About Cinéthos
          </span>
        </h1>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          A Hypertube-inspired web app for searching and watching movies with a clean, modern UX.
        </p>
      </header>

      {/* What */}
      <section className="rounded-2xl border border-gray-200 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-gray-800 dark:bg-gray-900/60">
        <h2 className="mb-2 text-xl font-semibold text-gray-900 dark:text-gray-100">What is this project?</h2>
        <p className="leading-relaxed text-gray-700 dark:text-gray-300">
          <strong>Cinéthos</strong> lets users <span className="font-semibold">search</span>, <span className="font-semibold">stream</span>, and
          <span className="font-semibold"> explore</span> movies directly in the browser. The player streams while the server downloads,
          ensuring minimal waiting time and an enjoyable experience.
        </p>
      </section>

      {/* Why */}
      <section className="rounded-2xl border border-gray-200 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-gray-800 dark:bg-gray-900/60">
        <h2 className="mb-2 text-xl font-semibold text-gray-900 dark:text-gray-100">Why we built it</h2>
        <p className="leading-relaxed text-gray-700 dark:text-gray-300">
          We wanted to craft a secure, performant, and elegant streaming experience using modern best
          practices—pushing ourselves across frontend, backend, and deployment. Cinéthos is our take
          on a production-ready, scalable app that respects the Hypertube subject rules.
        </p>
      </section>

      {/* Team */}
      <section className="rounded-2xl border border-gray-200 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-gray-800 dark:bg-gray-900/60">
        <h2 className="mb-3 text-xl font-semibold text-gray-900 dark:text-gray-100">Who works on it</h2>
        <ul className="grid grid-cols-1 gap-2">
          {["abouazi", "stamim", "anaji-el", "amessah"].map((name) => (
            <li
              key={name}
              className="flex items-center justify-between rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-gray-800 shadow-sm transition hover:shadow dark:border-gray-800 dark:bg-gray-800 dark:text-gray-100"
            >
              <span className="font-medium">@{name}</span>
              <span className="select-none rounded-full border border-gray-200 px-2 py-0.5 text-xs text-gray-500 dark:border-gray-700 dark:text-gray-400">
                Team Cinéthos
              </span>
            </li>
          ))}
        </ul>
      </section>

      {/* Tech Stack */}
      <section className="rounded-2xl border border-gray-200 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-gray-800 dark:bg-gray-900/60">
        <h2 className="mb-3 text-xl font-semibold text-gray-900 dark:text-gray-100">Technologies Used</h2>
        <p className="mb-3 text-gray-700 dark:text-gray-300">
          We use the latest tooling and stick to the subject’s requirements for search, streaming,
          security, and API design. Light &amp; dark modes are supported out of the box.
        </p>
        <ul className="flex flex-wrap gap-2">
          {[
            "Next.js (TypeScript)",
            "Tailwind CSS",
            "Spring Boot",
            "Go",
            "RESTful API + OAuth2",
            "OMDb / TMDb (per subject)",
          ].map((t) => (
            <li
              key={t}
              className="rounded-full border border-gray-200 bg-gray-50 px-3 py-1 text-sm text-gray-700 dark:border-gray-800 dark:bg-gray-800 dark:text-gray-200"
            >
              {t}
            </li>
          ))}
        </ul>
      </section>

      <section className="rounded-2xl border border-gray-200 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-gray-800 dark:bg-gray-900/60">
        <h2 className="mb-2 text-xl font-semibold text-gray-900 dark:text-gray-100">Subject Compliance</h2>
        <ul className="list-disc space-y-1 pl-5 text-gray-700 dark:text-gray-300">
          <li>Authenticated library with search and thumbnails (sorted, filterable, paginated).</li>
          <li>Video details page with player, metadata, comments, and on-the-fly streaming.</li>
          <li>Server-side download/cleanup policy and subtitles per user language when available.</li>
          <li>Secure forms, no plain-text passwords, input validation, and safe uploads.</li>
          <li>Public REST API with OAuth2 for users, movies, and comments (per spec).</li>
          <li>Responsive UI with accessible light/dark themes.</li>
        </ul>
      </section>

      {/* <footer className="pt-2 text-center text-xs text-gray-500 dark:text-gray-400">
        © {new Date().getFullYear()} {"Cinéthos".toUpperCase()} — built with care by the team.
      </footer> */}
    </article>
  );
}
