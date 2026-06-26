import { z } from "zod"

export const loginSchema = z.object({
  email: z.email("Enter a valid email"),
  password: z.string().min(1, "Password is required"),
})

export const registerSchema = z.object({
  email: z.email("Enter a valid email"),
  password: z.string()
    .min(8, "At least 8 characters")
    .regex(/[A-Z]/, "At least one uppercase letter")
    .regex(/[0-9]/, "At least one number"),
  confirmPassword: z.string(),
}).refine((d) => d.password === d.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
})

export const createLinkSchema = z.object({
  url: z.url("Enter a valid URL (must start with http:// or https://)"),
  alias: z
    .string()
    .regex(/^[a-zA-Z0-9_-]{3,50}$/, "3–50 chars, letters/digits/underscore/hyphen only")
    .optional()
    .or(z.literal("")),
  expires_at: z.string().optional(),
})

export type LoginInput = z.infer<typeof loginSchema>
export type RegisterInput = z.infer<typeof registerSchema>
export type CreateLinkInput = z.infer<typeof createLinkSchema>
