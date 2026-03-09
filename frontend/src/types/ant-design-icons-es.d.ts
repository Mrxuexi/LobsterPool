declare module '@ant-design/icons/es/icons/*' {
  import type { AntdIconProps } from '@ant-design/icons/es/components/AntdIcon';
  import type { ForwardRefExoticComponent, RefAttributes } from 'react';

  const Icon: ForwardRefExoticComponent<AntdIconProps & RefAttributes<HTMLSpanElement>>;

  export default Icon;
}
