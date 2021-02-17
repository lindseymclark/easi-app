import { gql } from '@apollo/client';

export default gql`
  query GetSystems($first: Int!) {
    systems(first: $first) {
      edges {
        node {
          id
          name
          businessOwner {
            name
            component
          }
        }
      }
    }
  }
`;