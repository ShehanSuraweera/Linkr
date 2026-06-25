"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { loginSchema, type LoginInput } from "@/lib/schemas"

export default function LoginPage() {
  const router = useRouter()
  const [serverError, setServerError] = useState("")
  const [isRegistering, setIsRegistering] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginInput>({ resolver: zodResolver(loginSchema) })

  const onSubmit = async (data: LoginInput) => {
    setServerError("")
    const endpoint = isRegistering ? "/api/auth/register" : "/api/auth/login"
    try {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      })
      const body = await res.json()
      if (!res.ok) {
        setServerError(body.error ?? "Something went wrong")
        return
      }
      router.push("/dashboard")
      router.refresh()
    } catch {
      setServerError("Network error — is the API running?")
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm bg-white rounded-2xl shadow-sm border border-gray-200 p-8">
        <h1 className="text-2xl font-bold mb-1">Linkr</h1>
        <p className="text-gray-500 text-sm mb-6">
          {isRegistering ? "Create an account" : "Sign in to your account"}
        </p>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input
              {...register("email")}
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
            {errors.email && (
              <p className="text-red-500 text-xs mt-1">{errors.email.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <input
              {...register("password")}
              type="password"
              autoComplete={isRegistering ? "new-password" : "current-password"}
              placeholder="••••••••"
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
            {errors.password && (
              <p className="text-red-500 text-xs mt-1">{errors.password.message}</p>
            )}
          </div>

          {serverError && (
            <p className="text-red-600 text-sm bg-red-50 rounded-lg px-3 py-2">{serverError}</p>
          )}

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
          >
            {isSubmitting ? "Please wait…" : isRegistering ? "Create account" : "Sign in"}
          </button>
        </form>

        <p className="text-center text-sm text-gray-500 mt-4">
          {isRegistering ? "Already have an account?" : "No account?"}{" "}
          <button
            onClick={() => { setIsRegistering(!isRegistering); setServerError("") }}
            className="text-indigo-600 font-medium hover:underline"
          >
            {isRegistering ? "Sign in" : "Register"}
          </button>
        </p>
      </div>
    </div>
  )
}
