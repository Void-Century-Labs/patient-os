"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { getQueueEntry, getQueueWsUrl, leaveQueue, type QueueUpdateMessage } from "@/lib/api";

const STATUS_LABEL: Record<string, string> = {
  waiting: "Waiting",
  called: "Your turn — please proceed",
  completed: "Consultation complete",
  skipped: "Skipped",
  cancelled: "Left queue",
};

export default function QueueTrackingPage() {
  const params = useParams<{ hospitalId: string; entryId: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const entryId = Number(params.entryId);
  const [wsConnected, setWsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  const { data: entry, isLoading, isError, error } = useQuery({
    queryKey: ["queue-entry", entryId],
    queryFn: () => getQueueEntry(entryId),
    refetchInterval: (query) =>
      !wsConnected && query.state.data?.status === "waiting" ? 5000 : false,
  });

  const queueId = entry?.queue_id;

  useEffect(() => {
    if (!queueId) return;

    const socket = new WebSocket(getQueueWsUrl(queueId));
    wsRef.current = socket;

    socket.onopen = () => setWsConnected(true);
    socket.onclose = () => setWsConnected(false);
    socket.onerror = () => setWsConnected(false);
    socket.onmessage = (event) => {
      const message: QueueUpdateMessage = JSON.parse(event.data);
      const updated = message.entries.find((item) => item.id === entryId);
      if (updated) {
        queryClient.setQueryData(["queue-entry", entryId], updated);
      } else {
        queryClient.invalidateQueries({ queryKey: ["queue-entry", entryId] });
      }
    };

    return () => {
      socket.close();
      wsRef.current = null;
    };
  }, [queueId, entryId, queryClient]);

  const leaveMutation = useMutation({
    mutationFn: () => leaveQueue(entryId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["queue-entry", entryId] });
    },
  });

  return (
    <main className="mx-auto flex w-full max-w-md flex-1 flex-col gap-6 px-6 py-16">
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Your queue status</h1>
        <p className="text-muted-foreground text-sm">
          This page updates automatically while you wait.
        </p>
      </div>

      {isLoading && <p className="text-muted-foreground text-center text-sm">Loading…</p>}
      {isError && <p className="text-center text-sm text-destructive">{error.message}</p>}

      {entry && (
        <Card>
          <CardHeader className="text-center">
            <CardDescription>Your token</CardDescription>
            <CardTitle className="text-5xl">#{entry.token_number}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4">
            <Badge
              variant={
                entry.status === "called"
                  ? "success"
                  : entry.status === "cancelled" || entry.status === "skipped"
                    ? "destructive"
                    : "secondary"
              }
              className="text-sm"
            >
              {STATUS_LABEL[entry.status] ?? entry.status}
            </Badge>

            {entry.status === "waiting" && (
              <div className="flex w-full justify-around text-center">
                <div>
                  <p className="text-2xl font-semibold">{entry.position}</p>
                  <p className="text-muted-foreground text-xs">Position in queue</p>
                </div>
                <div>
                  <p className="text-2xl font-semibold">{entry.estimated_wait_minutes}</p>
                  <p className="text-muted-foreground text-xs">Est. wait (min)</p>
                </div>
              </div>
            )}

            {entry.status === "waiting" && (
              <Button
                variant="outline"
                className="w-full"
                disabled={leaveMutation.isPending}
                onClick={() => leaveMutation.mutate()}
              >
                {leaveMutation.isPending ? "Leaving…" : "Leave Queue"}
              </Button>
            )}

            {entry.status === "cancelled" && (
              <Button
                className="w-full"
                onClick={() => router.push(`/h/${params.hospitalId}/discover`)}
              >
                Choose another doctor
              </Button>
            )}
          </CardContent>
        </Card>
      )}
    </main>
  );
}
