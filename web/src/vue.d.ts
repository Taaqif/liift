import Vue from "vue";

declare module "vue" {
  export interface AllowedComponentProps extends HTMLAttributes {
    onClick?: ((payload: MouseEvent) => void) | undefined;
    tabindex?: number | string | undefined;
  }

  export default Vue;
}
