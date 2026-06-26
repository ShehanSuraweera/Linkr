// components/LinkrLogoIcon.tsx
export default function LinkrLogoIcon() {
  return (
    <svg
      width="160"
      height="48"
      viewBox="0 0 160 48"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="Linkr Logo"
    >
      {/* Chain Icon */}
      <g transform="translate(4, 12)">
        <path
          d="M12 4H7C4.23858 4 2 6.23858 2 9C2 11.7614 4.23858 14 7 14H12"
          stroke="#D31E54"
          strokeWidth="3.5"
          strokeLinecap="round"
        />
        <path
          d="M10 14H15C17.7614 14 20 11.7614 20 9C20 6.23858 17.7614 4 15 4H10"
          stroke="currentColor"
          strokeWidth="3.5"
          strokeLinecap="round"
        />
        <line
          x1="7" y1="9" x2="15" y2="9"
          stroke="#D31E54"
          strokeWidth="3.5"
          strokeLinecap="round"
        />
      </g>

      {/* Text */}
      <text
        x="38"
        y="32"
        fontFamily="ui-sans-serif, system-ui, sans-serif"
        fontSize="26"
        fontWeight="800"
        fill="currentColor"
        letterSpacing="-1"
      >
        Linkr
      </text>
    </svg>
  );
}
