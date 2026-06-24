"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { getDiscovery, joinQueue } from "@/lib/api";
import { loadPatient } from "@/lib/patient-session";

export default function DiscoverPage() {
  const params = useParams<{ hospitalId: string }>();
  const router = useRouter();
  const [joiningDoctorId, setJoiningDoctorId] = useState<number | null>(null);

  const patient = loadPatient();

  useEffect(() => {
    if (!patient) {
      router.replace(`/h/${params.hospitalId}/register`);
    }
  }, [patient, params.hospitalId, router]);

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["discovery", params.hospitalId],
    queryFn: () => getDiscovery(params.hospitalId),
    enabled: !!patient,
  });

  const joinMutation = useMutation({
    mutationFn: (doctorId: number) => {
      setJoiningDoctorId(doctorId);
      return joinQueue(doctorId, patient!.id);
    },
    onSuccess: (entry) => {
      router.push(`/h/${params.hospitalId}/queue/${entry.id}`);
    },
    onSettled: () => setJoiningDoctorId(null),
  });

  if (!patient) return null;

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-6 py-16">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">Choose a doctor</h1>
        <p className="text-muted-foreground text-sm">
          Pick a department and doctor to join their queue.
        </p>
      </div>

      {isLoading && <p className="text-muted-foreground text-sm">Loading departments…</p>}
      {isError && <p className="text-sm text-destructive">{error.message}</p>}
      {joinMutation.isError && (
        <p className="text-sm text-destructive">{joinMutation.error.message}</p>
      )}

      {data?.length === 0 && (
        <p className="text-muted-foreground text-sm">No departments available yet.</p>
      )}

      <div className="space-y-4">
        {data?.map((department) => (
          <Card key={department.id}>
            <CardHeader>
              <CardTitle>{department.name}</CardTitle>
              <CardDescription>
                {department.doctors.length} doctor{department.doctors.length === 1 ? "" : "s"} available
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {department.doctors.map((doctor) => (
                <div
                  key={doctor.id}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="space-y-1">
                    <p className="font-medium">{doctor.name}</p>
                    <div className="flex items-center gap-2">
                      <Badge variant="secondary">Queue: {doctor.queue_length}</Badge>
                      <Badge variant="outline">~{doctor.estimated_wait_minutes} min wait</Badge>
                    </div>
                  </div>
                  <Button
                    size="sm"
                    disabled={joinMutation.isPending}
                    onClick={() => joinMutation.mutate(doctor.id)}
                  >
                    {joinMutation.isPending && joiningDoctorId === doctor.id
                      ? "Joining…"
                      : "Join Queue"}
                  </Button>
                </div>
              ))}
            </CardContent>
          </Card>
        ))}
      </div>
    </main>
  );
}
