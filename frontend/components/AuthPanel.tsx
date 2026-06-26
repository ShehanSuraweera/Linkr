import { Link2, BarChart2, Zap } from "lucide-react"
import LinkrLogoIcon from "./LinkrLogoIcon"

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
      // CHANGED: Deepened the gradient to a richer, darker maroon/rose for better contrast
      style={{ background: "linear-gradient(145deg, #310413 0%, #701133 55%, #9f1239 100%)" }}
    >
      <BackgroundDecoration />

      <LinkrLogoIcon size="lg"/>

      {/* Tagline */}
      <div className="relative z-10 flex-1 flex flex-col justify-center">
        <h1 className="text-5xl font-extrabold leading-[1.1] tracking-tight">
          Shorten.<br />Share.<br />Track.
        </h1>
        {/* CHANGED: Increased opacity from white/70 to white/90 */}
        <p className="mt-5 text-lg text-white/90 leading-relaxed max-w-xs">
          The simplest way to manage, share, and measure every link you create.
        </p>

        {/* Feature list */}
        <ul className="mt-10 space-y-4">
          {features.map(({ icon: Icon, label }) => (
            <li key={label} className="flex items-center gap-3">
              <div className="flex items-center justify-center w-8 h-8 rounded-full bg-white/20 shrink-0">
                <Icon className="h-4 w-4 text-white" />
              </div>
              {/* CHANGED: Increased opacity to white/95 for crisp readability */}
              <span className="text-sm font-medium text-white/95">{label}</span>
            </li>
          ))}
        </ul>

        {/* Mock dashboard card */}
        <div className="mt-10 rounded-2xl bg-white/10 border border-white/20 backdrop-blur-md overflow-hidden shadow-xl">
          <div className="px-4 py-3 border-b border-white/10 flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-white/60" />
            {/* CHANGED: Increased opacity from white/50 to white/70 */}
            <span className="text-xs text-white/70 font-mono">linkr.io</span>
          </div>
          {mockLinks.map(({ code, clicks }, i) => (
            <div
              key={code}
              className={`flex items-center justify-between px-4 py-3 ${i !== 0 ? "border-t border-white/10" : ""}`}
            >
              <span className="text-sm font-mono text-white/95">/{code}</span>
              {/* CHANGED: Increased opacity from white/50 to white/70 */}
              <span className="text-xs text-white/70">{clicks} clicks</span>
            </div>
          ))}
        </div>
      </div>

      {/* Footer */}
      {/* CHANGED: Increased opacity from white/35 to white/60. 35% is too low for WCAG accessibility */}
      <p className="relative z-10 text-xs text-white/60 mt-8">
        © {new Date().getFullYear()} Linkr. All rights reserved.
      </p>
    </div>
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
      <g fill="none" stroke="white" strokeWidth="28" opacity="0.05">
        <ellipse cx="340" cy="110" rx="90" ry="55" />
        <ellipse cx="470" cy="110" rx="90" ry="55" />
      </g>
      {/* Medium chain pair — centre left */}
      <g fill="none" stroke="white" strokeWidth="20" opacity="0.04">
        <circle cx="50"  cy="430" r="80" />
        <circle cx="175" cy="430" r="80" />
      </g>
      {/* Small chain pair — bottom right */}
      <g fill="none" stroke="white" strokeWidth="14" opacity="0.06">
        <ellipse cx="370" cy="690" rx="55" ry="36" />
        <ellipse cx="468" cy="690" rx="55" ry="36" />
      </g>
      {/* Glow blobs */}
      <circle cx="480" cy="0"   r="220" fill="white" opacity="0.02" />
      <circle cx="0"   cy="800" r="180" fill="white" opacity="0.03" />
      {/* Accent dots */}
      <circle cx="105" cy="210" r="4" fill="white" opacity="0.15" />
      <circle cx="355" cy="310" r="3" fill="white" opacity="0.10" />
      <circle cx="205" cy="615" r="5" fill="white" opacity="0.10" />
      <circle cx="405" cy="510" r="3" fill="white" opacity="0.12" />
    </svg>
  )
}