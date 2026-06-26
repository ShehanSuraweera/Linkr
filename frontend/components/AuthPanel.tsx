import { Link2, BarChart2, Zap } from "lucide-react"

const features = [
  { icon: Link2,     label: "Custom aliases & short codes" },
  { icon: BarChart2, label: "Click analytics & daily charts" },
  { icon: Zap,       label: "Fast, reliable redirects"       },
]

const mockLinks = [
  { code: "launch", clicks: "1,902" },
  { code: "docs",   clicks: "847"   },
  { code: "promo",  clicks: "431"   },
]

export default function AuthPanel() {
  return (
    <div
      className="relative hidden lg:flex flex-col h-full overflow-hidden p-12 text-white"
      style={{ background: "linear-gradient(145deg, #881337 0%, #be123c 55%, #e11d48 100%)" }}
    >
      <BackgroundDecoration />

      {/* Brand */}
      <div className="relative z-10 flex items-center gap-2.5">
        <ChainLinkIcon />
        <span className="text-2xl font-extrabold tracking-tight">Linkr</span>
      </div>

      {/* Tagline */}
      <div className="relative z-10 flex-1 flex flex-col justify-center">
        <h1 className="text-5xl font-extrabold leading-[1.1] tracking-tight">
          Shorten.<br />Share.<br />Track.
        </h1>
        <p className="mt-5 text-lg text-white/70 leading-relaxed max-w-xs">
          The simplest way to manage, share, and measure every link you create.
        </p>

        {/* Feature list */}
        <ul className="mt-10 space-y-4">
          {features.map(({ icon: Icon, label }) => (
            <li key={label} className="flex items-center gap-3">
              <div className="flex items-center justify-center w-8 h-8 rounded-full bg-white/15 shrink-0">
                <Icon className="h-4 w-4 text-white" />
              </div>
              <span className="text-sm font-medium text-white/85">{label}</span>
            </li>
          ))}
        </ul>

        {/* Mock dashboard card */}
        <div className="mt-10 rounded-2xl bg-white/10 border border-white/20 backdrop-blur-sm overflow-hidden">
          <div className="px-4 py-3 border-b border-white/10 flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-white/40" />
            <span className="text-xs text-white/50 font-mono">linkr.io</span>
          </div>
          {mockLinks.map(({ code, clicks }, i) => (
            <div
              key={code}
              className={`flex items-center justify-between px-4 py-3 ${i !== 0 ? "border-t border-white/10" : ""}`}
            >
              <span className="text-sm font-mono text-white/90">/{code}</span>
              <span className="text-xs text-white/50">{clicks} clicks</span>
            </div>
          ))}
        </div>
      </div>

      {/* Footer */}
      <p className="relative z-10 text-xs text-white/35 mt-8">
        © {new Date().getFullYear()} Linkr. All rights reserved.
      </p>
    </div>
  )
}

function ChainLinkIcon() {
  return (
    <svg width="26" height="26" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"
        stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
      />
      <path
        d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"
        stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
      />
    </svg>
  )
}

function BackgroundDecoration() {
  return (
    <svg
      className="absolute inset-0 w-full h-full"
      viewBox="0 0 480 800"
      preserveAspectRatio="xMidYMid slice"
      aria-hidden
    >
      {/* Large chain pair — top right */}
      <g fill="none" stroke="white" strokeWidth="28" opacity="0.07">
        <ellipse cx="340" cy="110" rx="90" ry="55" />
        <ellipse cx="470" cy="110" rx="90" ry="55" />
      </g>
      {/* Medium chain pair — centre left */}
      <g fill="none" stroke="white" strokeWidth="20" opacity="0.05">
        <circle cx="50"  cy="430" r="80" />
        <circle cx="175" cy="430" r="80" />
      </g>
      {/* Small chain pair — bottom right */}
      <g fill="none" stroke="white" strokeWidth="14" opacity="0.08">
        <ellipse cx="370" cy="690" rx="55" ry="36" />
        <ellipse cx="468" cy="690" rx="55" ry="36" />
      </g>
      {/* Glow blobs */}
      <circle cx="480" cy="0"   r="220" fill="white" opacity="0.03" />
      <circle cx="0"   cy="800" r="180" fill="white" opacity="0.04" />
      {/* Accent dots */}
      <circle cx="105" cy="210" r="4" fill="white" opacity="0.20" />
      <circle cx="355" cy="310" r="3" fill="white" opacity="0.15" />
      <circle cx="205" cy="615" r="5" fill="white" opacity="0.12" />
      <circle cx="405" cy="510" r="3" fill="white" opacity="0.18" />
    </svg>
  )
}
