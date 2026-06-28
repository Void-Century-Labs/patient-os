"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  createDepartment,
  createDoctor,
  getDepartments,
  getDoctors,
  type Department,
} from "@/lib/api";

function DepartmentCard({ hospitalId, department }: { hospitalId: string; department: Department }) {
  const [doctorName, setDoctorName] = useState("");
  const queryClient = useQueryClient();

  const { data: doctors, isLoading } = useQuery({
    queryKey: ["doctors", department.id],
    queryFn: () => getDoctors(department.id),
  });

  const createDoctorMutation = useMutation({
    mutationFn: (name: string) => createDoctor(department.id, name),
    onSuccess: () => {
      setDoctorName("");
      queryClient.invalidateQueries({ queryKey: ["doctors", department.id] });
    },
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>{department.name}</CardTitle>
        <CardDescription>
          {doctors?.length ?? 0} doctor{doctors?.length === 1 ? "" : "s"}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {isLoading && <p className="text-muted-foreground text-sm">Loading doctors…</p>}

        {doctors && doctors.length > 0 && (
          <div className="space-y-2">
            {doctors.map((doctor) => (
              <div key={doctor.id} className="flex items-center justify-between rounded-lg border p-3">
                <p className="font-medium">{doctor.name}</p>
                <Button asChild size="sm" variant="outline">
                  <Link href={`/h/${hospitalId}/operator/${doctor.id}`}>Open Dashboard</Link>
                </Button>
              </div>
            ))}
          </div>
        )}

        <form
          className="flex gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (doctorName.trim()) createDoctorMutation.mutate(doctorName.trim());
          }}
        >
          <Input
            placeholder="Doctor name"
            value={doctorName}
            onChange={(e) => setDoctorName(e.target.value)}
          />
          <Button type="submit" disabled={createDoctorMutation.isPending}>
            Add Doctor
          </Button>
        </form>
        {createDoctorMutation.isError && (
          <p className="text-sm text-destructive">{createDoctorMutation.error.message}</p>
        )}
      </CardContent>
    </Card>
  );
}

export default function AdminDashboardPage() {
  const params = useParams<{ hospitalId: string }>();
  const [departmentName, setDepartmentName] = useState("");
  const queryClient = useQueryClient();

  const { data: departments, isLoading, isError, error } = useQuery({
    queryKey: ["departments", params.hospitalId],
    queryFn: () => getDepartments(params.hospitalId),
  });

  const createDepartmentMutation = useMutation({
    mutationFn: (name: string) => createDepartment(params.hospitalId, name),
    onSuccess: () => {
      setDepartmentName("");
      queryClient.invalidateQueries({ queryKey: ["departments", params.hospitalId] });
    },
  });

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-16">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">Admin Dashboard</h1>
        <p className="text-muted-foreground text-sm">
          Manage departments and doctors for this hospital.
        </p>
      </div>

      {isLoading && <p className="text-muted-foreground text-sm">Loading departments…</p>}
      {isError && <p className="text-sm text-destructive">{error.message}</p>}

      <Card>
        <CardHeader>
          <CardTitle>Add Department</CardTitle>
        </CardHeader>
        <CardContent>
          <form
            className="flex gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              if (departmentName.trim()) createDepartmentMutation.mutate(departmentName.trim());
            }}
          >
            <div className="flex-1 space-y-2">
              <Label htmlFor="department-name">Department name</Label>
              <Input
                id="department-name"
                placeholder="Cardiology"
                value={departmentName}
                onChange={(e) => setDepartmentName(e.target.value)}
              />
            </div>
            <Button type="submit" className="self-end" disabled={createDepartmentMutation.isPending}>
              Add
            </Button>
          </form>
          {createDepartmentMutation.isError && (
            <p className="mt-2 text-sm text-destructive">{createDepartmentMutation.error.message}</p>
          )}
        </CardContent>
      </Card>

      {departments?.length === 0 && (
        <p className="text-muted-foreground text-sm">No departments yet. Add one above.</p>
      )}

      <div className="space-y-4">
        {departments?.map((department) => (
          <DepartmentCard key={department.id} hospitalId={params.hospitalId} department={department} />
        ))}
      </div>
    </main>
  );
}
