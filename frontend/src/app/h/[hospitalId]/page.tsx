import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { getHospital } from "@/lib/api";

export default async function HospitalLandingPage({
  params,
}: {
  params: Promise<{ hospitalId: string }>;
}) {
  const { hospitalId } = await params;
  const hospital = await getHospital(hospitalId).catch(() => null);

  return (
    <main className="mx-auto flex w-full max-w-md flex-1 flex-col items-center justify-center gap-6 px-6 py-16 text-center">
      <div className="space-y-2">
        <p className="text-sm text-muted-foreground">Welcome to</p>
        <h1 className="text-3xl font-semibold tracking-tight">
          {hospital?.name ?? "PatientOS"}
        </h1>
      </div>

      <Card className="w-full">
        <CardHeader>
          <CardTitle>Skip the line</CardTitle>
          <CardDescription>
            Register once, pick a doctor, and track your place in the queue
            from your phone.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button asChild className="w-full">
            <Link href={`/h/${hospitalId}/register`}>Get Started</Link>
          </Button>
        </CardContent>
      </Card>
    </main>
  );
}
