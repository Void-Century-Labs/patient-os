import Link from "next/link";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { getHospitals } from "@/lib/api";

export default async function Home() {
  const hospitals = await getHospitals().catch(() => []);

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-8 px-6 py-16">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">PatientOS</h1>
        <p className="text-muted-foreground">
          Hospital queue and patient navigation platform.
        </p>
      </div>

      {hospitals.length === 0 ? (
        <p className="text-muted-foreground text-sm">
          No hospitals found. Make sure the backend API is running and reachable.
        </p>
      ) : (
        <div className="space-y-4">
          <p className="text-muted-foreground text-sm">
            Select a hospital to register and join a queue. In production, patients
            reach this page directly via a hospital-specific QR code.
          </p>
          {hospitals.map((hospital) => (
            <Link key={hospital.id} href={`/h/${hospital.id}`}>
              <Card className="transition-colors hover:bg-accent">
                <CardHeader>
                  <CardTitle>{hospital.name}</CardTitle>
                  <CardDescription>Tap to get started</CardDescription>
                </CardHeader>
                <CardContent />
              </Card>
            </Link>
          ))}
        </div>
      )}
    </main>
  );
}
