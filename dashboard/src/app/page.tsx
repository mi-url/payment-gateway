import { redirect } from "next/navigation";

// Root page redirects to login.
// Authenticated users will be redirected to /dashboard by the auth middleware.
export default function Home() {
  redirect("/login");
}
