"use client"

import { useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { toast } from "sonner"
import { registerSchema, type RegisterInput } from "@/lib/schemas"
import { Eye, EyeOff } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import PasswordRules from "@/components/PasswordRules"
import { Check, X } from "lucide-react"
import AuthPanel from "@/components/AuthPanel"
import LinkrLogoIcon from "@/components/LinkrLogoIcon"

export default function RegisterPage() {
  const router = useRouter()
  const [serverError, setServerError] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [showRules, setShowRules] = useState(false)

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<RegisterInput>({ resolver: zodResolver(registerSchema) })

  const password = watch("password") ?? ""
  const confirmPassword = watch("confirmPassword") ?? ""

  const onSubmit = async (data: RegisterInput) => {
    setServerError("")
    try {
      const res = await fetch("/api/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: data.email, password: data.password }),
      })
      const body = await res.json()
      if (!res.ok) {
        setServerError(
          res.status === 409
            ? "An account with this email already exists"
            : body.error ?? "Could not create account"
        )
        return
      }
      toast.success("Account created!", { description: "Welcome to Linkr." })
      router.push("/dashboard")
      router.refresh()
    } catch {
      toast.error("Network error", { description: "Could not reach the API. Is the server running?" })
    }
  }

  return (
    <div className="min-h-screen grid lg:grid-cols-2">
      <AuthPanel />

      <div className="flex items-center justify-center px-8 py-12 bg-background">
        <div className="w-full max-w-sm">

          <div className="mb-8">
          <LinkrLogoIcon size="lg"/>
            <h2 className="text-2xl font-bold tracking-tight mt-6">Create an account</h2>
            <p className="text-muted-foreground text-sm mt-1">Start shortening and tracking your links</p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5" noValidate>

            {/* Email */}
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                {...register("email")}
                type="email"
                autoComplete="email"
                placeholder="you@example.com"
                aria-invalid={!!errors.email}
              />
              {errors.email && (
                <p className="text-destructive text-xs">{errors.email.message}</p>
              )}
            </div>

            {/* Password */}
            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <div className="relative">
                <Input
                  id="password"
                  {...register("password")}
                  type={showPassword ? "text" : "password"}
                  autoComplete="new-password"
                  placeholder="••••••••"
                  aria-invalid={!!errors.password}
                  className="pr-9"
                  onFocus={() => setShowRules(true)}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  aria-label={showPassword ? "Hide password" : "Show password"}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              {/* Live rule checklist replaces the static error message */}
              <PasswordRules password={password} show={showRules} />
            </div>

            {/* Confirm password */}
            <div className="space-y-1.5">
              <Label htmlFor="confirmPassword">Confirm password</Label>
              <div className="relative">
                <Input
                  id="confirmPassword"
                  {...register("confirmPassword")}
                  type={showConfirm ? "text" : "password"}
                  autoComplete="new-password"
                  placeholder="••••••••"
                  aria-invalid={!!errors.confirmPassword}
                  className="pr-9"
                  onPaste={(e) => e.preventDefault()}
                />
                <button
                  type="button"
                  onClick={() => setShowConfirm((v) => !v)}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  aria-label={showConfirm ? "Hide password" : "Show password"}
                >
                  {showConfirm ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              {confirmPassword.length > 0 ? (
                <p className={`flex items-center gap-1.5 text-xs ${confirmPassword === password ? "text-green-600 dark:text-green-400" : "text-destructive"}`}>
                  {confirmPassword === password
                    ? <><Check className="h-3.5 w-3.5 shrink-0" /> Passwords match</>
                    : <><X className="h-3.5 w-3.5 shrink-0" /> Passwords don&apos;t match</>
                  }
                </p>
              ) : (
                <p className="text-muted-foreground text-xs">Type your password again — pasting is disabled</p>
              )}
            </div>

            {serverError && (
              <p className="text-destructive text-sm bg-destructive/10 rounded-lg px-3 py-2">
                {serverError}
              </p>
            )}

            <Button type="submit" disabled={isSubmitting} className="w-full">
              {isSubmitting ? "Creating account…" : "Create account"}
            </Button>

          </form>

          <p className="text-center text-sm text-muted-foreground mt-6">
            Already have an account?{" "}
            <Link href="/login" className="text-primary font-medium hover:underline">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
