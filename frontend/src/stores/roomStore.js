import { create } from "zustand";

export const useRoomStore = create((set) => ({
  messages: [],
  addMessage: (msg) => set((state) => ({ messages: [...state.messages, msg] })),
  clearMessages: () => set({ messages: [] }),
  questionActive: false,
  winner: null,
  setQuestionActive: (active) => set({ questionActive: active }),
  setWinner: (name) => set({ winner: name }),
  readyStates: {},
  setReadyStates: (states) => set({ readyStates: states }),
  currentVideoId: null,
  setCurrentVideoId: (id) => set({ currentVideoId: id }),
}));
