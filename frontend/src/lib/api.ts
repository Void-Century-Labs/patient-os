const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080";

export type Hospital = {
  id: number;
  name: string;
};

export type Patient = {
  id: number;
  name: string;
  mobile: string;
  age: number;
  gender: string;
};

export type Doctor = {
  id: number;
  department_id: number;
  name: string;
};

export type Department = {
  id: number;
  hospital_id: number;
  name: string;
};

export type ActiveQueueEntry = QueueEntry & {
  patient_name: string;
};

export type DoctorAvailability = Doctor & {
  queue_length: number;
  estimated_wait_minutes: number;
};

export type DepartmentAvailability = {
  id: number;
  hospital_id: number;
  name: string;
  doctors: DoctorAvailability[];
};

export type QueueEntryStatus = "waiting" | "called" | "completed" | "skipped" | "cancelled";

export type QueueEntry = {
  id: number;
  queue_id: number;
  patient_id: number;
  token_number: number;
  status: QueueEntryStatus;
  position: number;
  estimated_wait_minutes: number;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `Request failed with status ${res.status}`);
  }

  return res.json() as Promise<T>;
}

export function getHospital(hospitalId: string) {
  return request<Hospital>(`/api/v1/hospitals/${hospitalId}`);
}

export function getHospitals() {
  return request<Hospital[]>("/api/v1/hospitals");
}

export function getDiscovery(hospitalId: string) {
  return request<DepartmentAvailability[]>(`/api/v1/hospitals/${hospitalId}/discovery`);
}

export function registerPatient(input: {
  name: string;
  mobile: string;
  age?: number;
  gender?: string;
}) {
  return request<Patient>("/api/v1/patients/register", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function joinQueue(doctorId: number, patientId: number) {
  return request<QueueEntry>(`/api/v1/doctors/${doctorId}/queue/join`, {
    method: "POST",
    body: JSON.stringify({ patient_id: patientId }),
  });
}

export function getQueueEntry(entryId: number) {
  return request<QueueEntry>(`/api/v1/queue-entries/${entryId}`);
}

export function leaveQueue(entryId: number) {
  return request<QueueEntry>(`/api/v1/queue-entries/${entryId}/leave`, {
    method: "POST",
  });
}

export type QueueUpdateMessage = {
  type: "queue_update";
  queue_id: number;
  entries: QueueEntry[];
};

export function getQueueWsUrl(queueId: number) {
  return `${WS_URL}/api/v1/queues/${queueId}/ws`;
}

export function getActiveQueue(doctorId: number) {
  return request<{ queue_id: number | null; entries: ActiveQueueEntry[] }>(
    `/api/v1/doctors/${doctorId}/queue`,
  );
}

export function callNext(doctorId: number) {
  return request<QueueEntry>(`/api/v1/doctors/${doctorId}/queue/call-next`, {
    method: "POST",
  });
}

export function completeEntry(entryId: number) {
  return request<QueueEntry>(`/api/v1/queue-entries/${entryId}/complete`, {
    method: "POST",
  });
}

export function skipEntry(entryId: number) {
  return request<QueueEntry>(`/api/v1/queue-entries/${entryId}/skip`, {
    method: "POST",
  });
}

export function getDepartments(hospitalId: string) {
  return request<Department[]>(`/api/v1/hospitals/${hospitalId}/departments`);
}

export function createDepartment(hospitalId: string, name: string) {
  return request<Department>(`/api/v1/hospitals/${hospitalId}/departments`, {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export function getDoctors(departmentId: number) {
  return request<Doctor[]>(`/api/v1/departments/${departmentId}/doctors`);
}

export function createDoctor(departmentId: number, name: string) {
  return request<Doctor>(`/api/v1/departments/${departmentId}/doctors`, {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export type Notification = {
  id: number;
  patient_id: number;
  type: string;
  message: string;
  created_at: string;
};

export function getNotifications(patientId: number) {
  return request<Notification[]>(`/api/v1/patients/${patientId}/notifications`);
}
