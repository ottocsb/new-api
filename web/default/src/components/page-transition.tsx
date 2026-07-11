import { Outlet, useRouterState } from '@tanstack/react-router'
import { motion, useReducedMotion, type Variants } from 'motion/react'
import type { ReactNode } from 'react'

import { MOTION_TRANSITION, MOTION_VARIANTS } from '@/lib/motion'

interface PageTransitionProps {
  children: ReactNode
  className?: string
}

export function PageTransition(props: PageTransitionProps) {
  const shouldReduce = useReducedMotion()

  if (shouldReduce) {
    return <div className={props.className}>{props.children}</div>
  }

  return (
    <motion.div
      initial={MOTION_VARIANTS.pageEnter.initial}
      animate={MOTION_VARIANTS.pageEnter.animate}
      transition={MOTION_TRANSITION.default}
      className={props.className}
    >
      {props.children}
    </motion.div>
  )
}

export function AnimatedOutlet() {
  const shouldReduce = useReducedMotion()
  // Key the page transition by the matched route id, not the resolved pathname.
  // Navigating between params of the same route (e.g. dashboard tabs served by
  // /dashboard/$section) then re-renders in place instead of remounting the
  // route component and discarding its state (such as the selected time range).
  const routeKey = useRouterState({
    select: (s) => s.matches.at(-1)?.routeId ?? s.location.pathname,
  })

  if (shouldReduce) {
    return (
      <div className='flex min-h-0 flex-1 flex-col'>
        <Outlet />
      </div>
    )
  }

  return (
    <motion.div
      key={routeKey}
      initial={MOTION_VARIANTS.pageEnter.initial}
      animate={MOTION_VARIANTS.pageEnter.animate}
      transition={MOTION_TRANSITION.fast}
      className='flex min-h-0 flex-1 flex-col'
    >
      <Outlet />
    </motion.div>
  )
}

interface StaggerContainerProps {
  children: ReactNode
  className?: string
  variants?: Variants
}

/* Entrance choreography is retired on admin surfaces: content appears
 * immediately so tables and dashboards read as instant. The components stay
 * as plain wrappers so call sites keep working; decorative motion remains
 * available to the landing page via its dedicated landing-* utilities. */

export function StaggerContainer(props: StaggerContainerProps) {
  return <div className={props.className}>{props.children}</div>
}

interface StaggerItemProps {
  children: ReactNode
  className?: string
  variants?: Variants
}

export function StaggerItem(props: StaggerItemProps) {
  return <div className={props.className}>{props.children}</div>
}

export function TableStaggerContainer(props: StaggerContainerProps) {
  return <tbody className={props.className}>{props.children}</tbody>
}

export function TableStaggerRow(props: StaggerItemProps) {
  return <tr className={props.className}>{props.children}</tr>
}

export function CardStaggerContainer(props: StaggerContainerProps) {
  return <div className={props.className}>{props.children}</div>
}

export function CardStaggerItem(props: StaggerItemProps) {
  return <div className={props.className}>{props.children}</div>
}

interface FadeInProps {
  children: ReactNode
  className?: string
  delay?: number
}

export function FadeIn(props: FadeInProps) {
  const shouldReduce = useReducedMotion()

  if (shouldReduce) {
    return <div className={props.className}>{props.children}</div>
  }

  return (
    <motion.div
      initial={MOTION_VARIANTS.fadeIn.initial}
      animate={MOTION_VARIANTS.fadeIn.animate}
      transition={{
        ...MOTION_TRANSITION.default,
        delay: props.delay,
      }}
      className={props.className}
    >
      {props.children}
    </motion.div>
  )
}
