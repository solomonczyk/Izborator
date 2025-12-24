# DEVELOPMENT FLOW — Пошаговый процесс разработки

## 1. Стадии разработки каждой фичи
Каждая фича проходит стадии:

1. Design  
2. Skeleton  
3. Implementation  
4. Integration  
5. Testing  
6. Logging in DEVELOPMENT_LOG.md  

## 2. Cursor и AI обязаны:
- читать ROADMAP_CURRENT_STEP.md перед каждым действием
- выполнять только текущий шаг
- запрашивать уточнение, если шаг не ясен

## 3. Нельзя перепрыгивать стадии
Implementation невозможна до завершения Design.

## 4. После каждого шага:
- обновить DEVELOPMENT_LOG.md
- обновить STATUS.md
- обновить ROADMAP_CURRENT_STEP.md (если шаг завершён)

## 5. Минимальные изменения
ИИ должен менять код минимально, строго в рамках задачи.
