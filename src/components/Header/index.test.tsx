import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import { shallow, mount } from 'enzyme';
import { act } from 'react-dom/test-utils';

import { Header } from './index';

describe('The Header component', () => {
  jest.mock('@okta/okta-react', () => {
    console.log('am i executed');
    return {
      useOktaAuth: () => ({
        authState: {},
        authService: {}
      })
    };
  });

  xit('renders without crashing', () => {
    shallow(<Header />);
  });

  describe('When logged out', () => {
    const auth = {
      isAuthenticated: () => Promise.resolve(false),
      user: {
        name: ''
      }
    };
    xit('displays a login button', () => {
      const component = shallow(<Header auth={auth} />);
      expect(component.text().includes('Login')).toBe(true);
      expect(component.text().includes('Logout')).toBe(false);
    });
  });

  describe('When logged in', () => {
    const auth = {
      isAuthenticated: () => Promise.resolve(true),
      getUser: () =>
        Promise.resolve({
          name: 'John Doe'
        })
    };

    xit('displays a login button', async done => {
      let component;
      await act(async () => {
        component = mount(
          <BrowserRouter>
            <Header auth={auth} />
          </BrowserRouter>
        );
      });

      setImmediate(() => {
        component.update();
        expect(component.text().includes('Logout')).toBe(true);
        expect(component.text().includes('Login')).toBe(false);
        done();
      });
    });

    xit('displays the users name', async done => {
      let component;

      await act(async () => {
        component = mount(
          <BrowserRouter>
            <Header auth={auth} />
          </BrowserRouter>
        );
      });

      setImmediate(() => {
        component.update();
        expect(component.text().includes('John Doe')).toBe(true);
        done();
      });
    });

    xit('displays dropdown when caret is clicked', async done => {
      let component;

      await act(async () => {
        component = mount(
          <BrowserRouter>
            <Header auth={auth} />
          </BrowserRouter>
        );
      });

      setImmediate(() => {
        component.update();
        expect(component.find('.user-actions-dropdown').exists()).toBe(false);
        component.find('.easi-header__caret').simulate('click');
        expect(component.find('.user-actions-dropdown').exists()).toBe(true);
        done();
      });
    });
  });

  xit('displays children', () => {
    const auth = {
      isAuthenticated: () => Promise.resolve(true),
      user: {
        name: ''
      }
    };

    const component = shallow(
      <Header auth={auth}>
        <div className="test-class-name" />
      </Header>
    );

    expect(component.find('.test-class-name').exists()).toBe(true);
  });
});
