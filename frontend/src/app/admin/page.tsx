"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";

import { HospitalQrCode } from "@/components/hospital-qr-code";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createHospital, getHospitalUrl, getHospitals } from "@/lib/api";

export default function HospitalAdminPage() {
  const [name, setName] = useState("");
  const queryClient = useQueryClient();

  const { data: hospitals, isLoading, isError, error } = useQuery({
    queryKey: ["hospitals"],
    queryFn: getHospitals,
  });

  const createHospitalMutation = useMutation({
    mutationFn: (name: string) => createHospital(name),
    onSuccess: () => {
      setName("");
      queryClient.invalidateQueries({ queryKey: ["hospitals"] });
    },
  });

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-16">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">Hospital Network Admin</h1>
        <p className="text-muted-foreground text-sm">
          Add new hospitals and get a QR code patients can scan to register.
        </p>
      </div>

      {isLoading && <p className="text-muted-foreground text-sm">Loading hospitals…</p>}
      {isError && <p className="text-sm text-destructive">{error.message}</p>}

      <Card>
        <CardHeader>
          <CardTitle>Add Hospital</CardTitle>
        </CardHeader>
        <CardContent>
          <form
            className="flex gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              if (name.trim()) createHospitalMutation.mutate(name.trim());
            }}
          >
            <div className="flex-1 space-y-2">
              <Label htmlFor="hospital-name">Hospital name</Label>
              <Input
                id="hospital-name"
                placeholder="City General Hospital"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
            <Button type="submit" className="self-end" disabled={createHospitalMutation.isPending}>
              Add
            </Button>
          </form>
          {createHospitalMutation.isError && (
            <p className="mt-2 text-sm text-destructive">{createHospitalMutation.error.message}</p>
          )}
        </CardContent>
      </Card>

      {hospitals?.length === 0 && (
        <p className="text-muted-foreground text-sm">No hospitals yet. Add one above.</p>
      )}

      <div className="space-y-4">
        {hospitals?.map((hospital) => (
          <Card key={hospital.id}>
            <CardHeader>
              <CardTitle>{hospital.name}</CardTitle>
              <CardDescription>{getHospitalUrl(hospital.id)}</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col items-start gap-4 sm:flex-row sm:items-center sm:justify-between">
              <HospitalQrCode value={getHospitalUrl(hospital.id)} size={150} />
              <Button asChild variant="outline">
                <Link href={`/h/${hospital.id}/admin`}>Manage Departments & Doctors</Link>
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
    </main>
  );
}
