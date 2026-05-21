import api from './client';
import type { Circle, CircleMember } from '../types';

export async function listCircles(): Promise<Circle[]> {
  const res = await api.get<Circle[]>('/api/circles/');
  return res.data;
}

export async function createCircle(data: {
  name: string;
  description: string;
  type: 'personal' | 'family' | 'group';
}): Promise<Circle> {
  const res = await api.post<Circle>('/api/circles/', data);
  return res.data;
}

export async function getCircle(id: string): Promise<Circle> {
  const res = await api.get<Circle>(`/api/circles/${id}`);
  return res.data;
}

export async function getCircleMembers(id: string): Promise<CircleMember[]> {
  const res = await api.get<CircleMember[]>(`/api/circles/${id}/members`);
  return res.data;
}

export async function deleteCircle(id: string): Promise<void> {
  await api.delete(`/api/circles/${id}`);
}
