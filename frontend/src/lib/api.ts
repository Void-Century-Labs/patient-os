const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

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
