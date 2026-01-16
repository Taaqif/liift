import Vue from "vue";

declare module "vue" {
  export interface AllowedComponentProps extends HTMLAttributes {
    onClick?: ((payload: MouseEvent) => void) | undefined;
  }

  export default Vue;
}
