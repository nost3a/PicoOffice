<template>
  <!-- touch-device top-right badge, marks touch-optimized build-->
  <div v-if="isTouch" class="touch-badge">📱 {{ $t('touch.badge') }}</div>
</template>

<script setup>
import { useI18n } from '@/i18n'
import { useTouch } from './useTouch'
const { isTouch } = useTouch()
useI18n()
</script>

<style scoped>
.touch-badge {
  position: fixed;
  right: 10px;
  bottom: 10px;
  z-index: 999;
  background: rgba(64, 158, 255, 0.9);
  color: #fff;
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
</style>
