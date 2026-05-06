import { createClient } from '@supabase/supabase-js'

const supabaseUrl = 'https://pceqzxcyrfcwrascnawy.supabase.co'
const supabaseAnonKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InBjZXF6eGN5cmZjd3Jhc2NuYXd5Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzY4ODM3ODUsImV4cCI6MjA5MjQ1OTc4NX0.91Z7TuEMzTttvAFRZpp4HS5B7ayualYkgPwuN6Q8zrY'

// Note: Using service role key for storage operations if needed, but anon is usually enough for uploads if bucket is public/configured
// However, the user provided the service key so I will use it for administrative tasks if necessary.
// For standard client-side uploads, anon is safer.
export const supabase = createClient(supabaseUrl, supabaseAnonKey)

// Service role key for administrative tasks (bypassing RLS)
const supabaseServiceKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InBjZXF6eGN5cmZjd3Jhc2NuYXd5Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc3Njg4Mzc4NSwiZXhwIjoyMDkyNDU5Nzg1fQ.0hH2RF3E-2LXbL-7wqGTOolgYuXikGPyEwU5P9gBGMQ'
export const supabaseAdmin = createClient(supabaseUrl, supabaseServiceKey)
