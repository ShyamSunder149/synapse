# Contributing to Synapse Framework

The project welcomes contributions from the community!

## Set up your development environment

1. Install [Go](https://golang.org/) (if not already installed)

2. Clone the repository:

   ```
   git clone https://github.com/ritvikos/synapse
   ```

3. Navigate to the project directory:

   ```
   cd synapse
   ```

4. [Install Taskfile](https://taskfile.dev/docs/installation#official-package-managers) (if not already installed)

5. Run all tests to verify your setup with `task test:all` or run individual component tests with `task test PKG=fetcher` (replace `fetcher` with the desired component name, see [project's Taskfile](./Taskfile.yml)).

## Considerations for designing abstractions and APIs

1. Try to stay compatible with standard library interfaces when applicable..
2. Prefer interface segregation for low couping, when applicable with bare-minimum, necessary methods, when possible.
3. Try to expose safe minimal public API surface with sensible configurable options and defaults. In case the unsafe operations are exposed for standard library compatibility, document them clearly about the potential risks and usage guidelines.

## Contributing

Before writing code for a new feature, please first discuss the change you wish to make via github discussions to ensure that it aligns with the project's goals and to avoid any duplication of effort.

To contribute, please follow these guidelines:

1. Fork the repository: [https://github.com/ritvikos/synapse](https://github.com/ritvikos/synapse)

2. Clone your forked repository to your local machine.

```
git clone forked-repo-url
```

3. Create a new branch for your feature or bug fix.

```
git checkout -b your-bug-or-feature-branch
```

4. Once you have made your changes, ensure that you:

   - Add/Modify tests to cover your changes
   - Update relevant documentation
   - Add/Modify examples, if necessary
   - Run the test suite with `task test` to verify everything works as expected

5. Run `task lint` to ensure your code adheres to the project's coding standards

6. Make your changes and commit them with clear, [descriptive conventoinal commit message](https://www.conventionalcommits.org/en/v1.0.0/). Ensure that you've [\"Sign-off\" your commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits). You can do this by adding the `-s` flag to your git commit command:

   ```
   git commit -s
   ```

7. Push your changes to your forked repository.

8. Open a Pull Request (PR) against the relevant branch in the Synapse repository with a clear description of your changes and reference any related issues, if applicable.
