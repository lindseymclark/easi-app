/* eslint-disable react/jsx-props-no-spreading */
import React from 'react';
import { shallow, mount } from 'enzyme';
import TextField from './index';

describe('The Text Field component', () => {
  const requiredProps = {
    id: 'DemoTest',
    name: 'Demo Input',
    onChange: () => {},
    onBlur: () => {},
    value: ''
  };

  it('renders without crashing', () => {
    shallow(<TextField {...requiredProps} />);
  });

  it('renders a label when provided', () => {
    const fixture = 'Demo Label';
    const component = shallow(<TextField {...requiredProps} label={fixture} />);

    expect(component.find('label').text()).toEqual(fixture);
  });

  it('triggers onChange', () => {
    const event = {
      target: {
        value: 'Hello'
      }
    };
    const mock = jest.fn();
    const component = mount(<TextField {...requiredProps} onChange={mock} />);

    component.simulate('change', event);
    expect(mock).toHaveBeenCalled();
  });
});