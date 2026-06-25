import { z } from "zod"

export const loginSchema = z.object({
  email: z.string().email("Enter a valid email"),
  password: z.string().min(8, "Password must be at least 8 characters"),
})

export const createLinkSchema = z.object({
  url: z.string().url("Enter a valid URL (must start with http:// or https://)"),
  alias: z
    .string()
    .regex(/^[a-zA-Z0-9_-]{3,50}$/, "3–50 chars, letters/digits/underscore/hyphen only")
    .optional()
    .or(z.literal("")),
  expires_at: z.string().optional(),
})

export type LoginInput = z.infer<typeof loginSchema>
export type CreateLinkInput = z.infer<typeof createLinkSchema>
