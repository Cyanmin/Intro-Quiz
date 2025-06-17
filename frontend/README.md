# Frontend Structure

This directory contains the React + Vite project for the Intro Quiz application.

## Directories

- `features/` - Feature specific components such as room or quiz modules.
- `components/` - Reusable UI components like buttons and inputs.
- `pages/` - Components used as routing pages.
- `hooks/` - Reusable React hooks.
- `services/` - API and WebSocket communication utilities.
- `stores/` - Global state management stores using Zustand.
- `utils/` - Generic utility functions.
- `routes/` - React Router configuration.

A minimal example of the Room feature is implemented to demonstrate the structure. The join screen connects to the backend WebSocket when the user joins a room.
