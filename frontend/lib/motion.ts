export const MOTION_TOKENS = {
  duration: {
    instant: "80ms",
    fast: "120ms",
    base: "160ms",
    slow: "220ms",
  },
  ease: {
    outSoft: "cubic-bezier(0.16, 1, 0.3, 1)",
    outBase: "cubic-bezier(0.2, 0, 0, 1)",
    inOut: "cubic-bezier(0.4, 0, 0.2, 1)",
  },
  amplitude: {
    moveXs: 4,
    moveSm: 8,
    moveMd: 12,
    scaleHover: 1.03,
    scalePress: 0.98,
    rotateXs: 2,
    rotateSm: 4,
  },
} as const;

export function prefersReducedMotion(): boolean {
  if (typeof window === "undefined") {
    return false;
  }

  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}
