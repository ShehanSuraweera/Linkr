import Image from "next/image";

const SIZES = {
  sm:  { withText: [80,  77],  icon: [28, 28] },
  md:  { withText: [116, 112], icon: [40, 40] },
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
      priority
    />
  );
}
