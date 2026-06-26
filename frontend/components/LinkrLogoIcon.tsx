import Image from "next/image";

// withText SVG is 2000×750 (8:3). icon SVG is 2000×2000 (1:1).
const SIZES = {
  sm:  { withText: [80,  30], icon: [28, 28] },
  md:  { withText: [116, 44], icon: [40, 40] },
  lg:  { withText: [300, 113], icon: [56, 56] },
} as const

interface LinkrLogoIconProps {
  withText?: boolean;
  size?: keyof typeof SIZES;
}

export default function LinkrLogoIcon({ withText = true, size = "md" }: LinkrLogoIconProps) {
  const [w, h] = withText ? SIZES[size].withText : SIZES[size].icon
  return (
    <Image
      src={withText ? "/images/logo-withtext.svg" : "/images/logo-withouttext.svg"}
      alt="Linkr"
      width={w}
      height={h}
      style={{ height: "auto" }}
      priority
    />
  );
}
