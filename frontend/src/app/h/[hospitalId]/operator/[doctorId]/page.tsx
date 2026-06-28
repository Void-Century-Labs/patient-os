"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { callNext, completeEntry, getActiveQueue, skipEntry } from "@/lib/api";

export default function OperatorDashboardPage() {
  const params = useParams<{ hospitalId: string; doctorId: string }>();
  const doctorId = Number(params.doctorId);
  const queryClient = useQueryClient();

  const queryKey = ["active-queue", doctorId];

  const { data, isLoading, isError, error } = useQuery({
    queryKey,
    queryFn: () => getActiveQueue(doctorId),
    refetchInterval: 5000,
  });

  const invalidate = () => queryClient.invalidateQueries({ queryKey });

  const callNextMutation = useMutation({
    mutationFn: () => callNext(doctorId),
    onSuccess: invalidate,
  });

  const completeMutation = useMutation({
    mutationFn: (entryId: number) => completeEntry(entryId),
    onSuccess: invalidate,
  });

  const skipMutation = useMutation({
    mutationFn: (entryId: number) => skipEntry(entryId),
    onSuccess: invalidate,
  });

  const entries = data?.entries ?? [];
  const hasWaiting = entries.some((entry) => entry.status === "waiting");

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-16">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">Operator Dashboard</h1>
        <p className="text-muted-foreground text-sm">
          Manage the live queue for this doctor.
        </p>
      </div>

      {isLoading && <p className="text-muted-foreground text-sm">Loading queue…</p>}
      {isError && <p className="text-sm text-destructive">{error.message}</p>}
      {callNextMutation.isError && (
        <p className="text-sm text-destructive">{callNextMutation.error.message}</p>
      )}

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Current Queue</CardTitle>
            <CardDescription>{entries.length} active patient{entries.length === 1 ? "" : "s"}</CardDescription>
          </div>
          <Button
            disabled={!hasWaiting || callNextMutation.isPending}
            onClick={() => callNextMutation.mutate()}
          >
            {callNextMutation.isPending ? "Calling…" : "Call Next"}
          </Button>
        </CardHeader>
        <CardContent>
          {entries.length === 0 ? (
            <p className="text-muted-foreground text-sm">No patients waiting.</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Token</TableHead>
                  <TableHead>Patient</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {entries.map((entry) => (
                  <TableRow key={entry.id}>
                    <TableCell>#{entry.token_number}</TableCell>
                    <TableCell>{entry.patient_name}</TableCell>
                    <TableCell>
                      <Badge variant={entry.status === "called" ? "success" : "secondary"}>
                        {entry.status === "called" ? "Called" : "Waiting"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        {entry.status === "called" && (
                          <Button
                            size="sm"
                            disabled={completeMutation.isPending}
                            onClick={() => completeMutation.mutate(entry.id)}
                          >
                            Complete
                          </Button>
                        )}
                        <Button
                          size="sm"
                          variant="outline"
                          disabled={skipMutation.isPending}
                          onClick={() => skipMutation.mutate(entry.id)}
                        >
                          Skip
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </main>
  );
}
