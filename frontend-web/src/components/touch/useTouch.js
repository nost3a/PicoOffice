// touch util: detect touch + narrow screen, return reactive flags
import { ref, onMounted, onBeforeUnmount } from 'vue'

const isTouch = ref(false)
const isNarrow = ref(false)
let inited = false

function detect() {
  isTouch.value =
    'ontouchstart' in window || (navigator.maxTouchPoints || 0) > 0
  isNarrow.value = window.innerWidth < 768
}

// back-key: if history stack only has this page, go to dashboard to avoid WebView exit
function onPopstate() {
  // let pages call router.back(); this is just a global listener stub
}

export function useTouch() {
  if (!inited) {
    inited = true
    onMounted(() => {
      detect()
      window.addEventListener('resize', detect)
      window.addEventListener('popstate', onPopstate)
    })
    onBeforeUnmount(() => {
      window.removeEventListener('resize', detect)
      window.removeEventListener('popstate', onPopstate)
    })
  }
  return { isTouch, isNarrow, detect }
}
