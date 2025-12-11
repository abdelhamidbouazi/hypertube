import { SVGProps } from "react";

export type IconSvgProps = SVGProps<SVGSVGElement> & {
  size?: number;
};

export type User = {
  id?: number;
  username?: string;
  firstname: string;
  lastname: string;
  email: string;
  avatar?: string;
  preferred_language?: string;
};

export type ContinueWatchingMovie = {
  id: number;
  title: string;
  posterPath?: string;
  genre: string;
  year: number;
};
