/*
Copyright (c) 2025 Tethys Plex

This file is part of Veloera.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
import React from 'react';
import { Icon } from '@douyinfe/semi-ui';

const NodelocIcon = (props) => {
  function CustomIcon() {
    return (
      <svg
        className='icon'
        viewBox='0 0 16 16'
        version='1.1'
        xmlns='http://www.w3.org/2000/svg'
        width='1em'
        height='1em'
        {...props}
      >
        <g id='nodeloc_icon' data-name='nodeloc_icon'>
          <circle cx='8' cy='8' r='8' fill='#2563eb' />
          <path
            d='M4 6h8v1H4V6zm0 2h8v1H4V8zm0 2h6v1H4v-1z'
            fill='white'
          />
          <circle cx='11' cy='5' r='1' fill='white' />
        </g>
      </svg>
    );
  }

  return <Icon svg={<CustomIcon />} />;
};

export default NodelocIcon;