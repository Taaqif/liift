import { ref, watch, onUnmounted } from "vue";

export function formatElapsed(sec: number): string {
  const s = Math.max(0, sec);
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const rem = s % 60;
  if (h > 0) return `${h}:${m.toString().padStart(2, "0")}:${rem.toString().padStart(2, "0")}`;
  return `${m}:${rem.toString().padStart(2, "0")}`;
}

export function useElapsedTimer(startedAt: () => string | null | undefined) {
  const elapsedSeconds = ref(0);
  let timerId: ReturnType<typeof setInterval> | null = null;

  function tick() {
    const at = startedAt();
    if (!at) { elapsedSeconds.value = 0; return; }
    elapsedSeconds.value = Math.max(0, Math.floor((Date.now() - new Date(at).getTime()) / 1000));
  }

  watch(
    startedAt,
    (at) => {
      if (timerId) { clearInterval(timerId); timerId = null; }
      if (at) {
        tick();
        timerId = setInterval(tick, 1000);
      } else {
        elapsedSeconds.value = 0;
      }
    },
    { immediate: true },
  );

  onUnmounted(() => { if (timerId) clearInterval(timerId); });

  return { elapsedSeconds };
}
