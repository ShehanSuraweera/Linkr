import LinkrLogoIcon from "@/components/LinkrLogoIcon"

export default function Footer() {
  return (
    <footer className="border-t bg-background mt-auto">
      <div className="max-w-5xl mx-auto px-6 py-5 flex flex-col sm:flex-row items-center justify-between gap-3">
        <LinkrLogoIcon />
        <p className="text-xs text-muted-foreground">
          &copy; {new Date().getFullYear()} Linkr. All rights reserved.
        </p>
      </div>
    </footer>
  )
}
