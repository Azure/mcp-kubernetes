from .command import ShellProcess
from .security_validator import validate_namespace_scope


def _helm(command_prefix: str, args: str) -> str:
    """
    Run a generic helm command and return the output.

    Args:
        command_prefix (str): The complete helm command prefix, e.g., 'helm list'.
        args (str): Arguments to pass to the command.

    Returns:
        str: The output of the helm command or an error message.
    """
    error = validate_namespace_scope(args)
    if error:
        return error

    process = ShellProcess(command=command_prefix)
    output = process.run(args)
    return output


# ----- Helm Read-Only Commands -----


def helm_list(args: str) -> str:
    """
    Run a `helm list` command and return the output.

    Args:
        args (str): The specific options to pass to `helm list`.
                       For example:
                       - "" (empty string) to list releases in the current namespace.
                       - "-n nginx-system" to list releases in a specific namespace.
                       - "-A" to list releases across all namespaces.

    Returns:
        str: The output of the `helm list` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm list` prefix; it is added automatically.
    """
    return _helm("helm list", args)


def helm_get(args: str) -> str:
    """
    Run a `helm get` command and return the output.

    Args:
        args (str): The specific subcommand and options to pass to `helm get`.
                       For example:
                       - "values release-name" to get values for a specific release.
                       - "manifest release-name" to get manifests for a release.
                       - "notes release-name" to get notes for a release.

    Returns:
        str: The output of the `helm get` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm get` prefix; it is added automatically.
    """
    return _helm("helm get", args)


def helm_status(args: str) -> str:
    """
    Run a `helm status` command and return the output.

    Args:
        args (str): The release name and options to pass to `helm status`.
                       For example:
                       - "release-name" to get status of a specific release.
                       - "release-name -n nginx-system" to get status in a specific namespace.

    Returns:
        str: The output of the `helm status` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm status` prefix; it is added automatically.
    """
    return _helm("helm status", args)


def helm_history(args: str) -> str:
    """
    Run a `helm history` command and return the output.

    Args:
        args (str): The release name and options to pass to `helm history`.
                       For example:
                       - "release-name" to get history of a specific release.
                       - "release-name -n nginx-system" to get history in a specific namespace.

    Returns:
        str: The output of the `helm history` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm history` prefix; it is added automatically.
    """
    return _helm("helm history", args)


def helm_search(args: str) -> str:
    """
    Run a `helm search` command and return the output.

    Args:
        args (str): The specific subcommand and options to pass to `helm search`.
                       For example:
                       - "repo nginx" to search for "nginx" in repositories.
                       - "hub prometheus" to search for "prometheus" in Helm Hub.

    Returns:
        str: The output of the `helm search` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm search` prefix; it is added automatically.
    """
    return _helm("helm search", args)


def helm_template(args: str) -> str:
    """
    Run a `helm template` command and return the output.

    Args:
        args (str): The release name, chart, and options to pass to `helm template`.
                       For example:
                       - "release-name chart-name" to render templates locally.
                       - "release-name chart-name --set key=value" to render with custom values.

    Returns:
        str: The output of the `helm template` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm template` prefix; it is added automatically.
        - This is a read-only command that renders templates locally and doesn't modify the cluster.
    """
    return _helm("helm template", args)


def helm_show(args: str) -> str:
    """
    Run a `helm show` command and return the output.

    Args:
        args (str): The specific subcommand and options to pass to `helm show`.
                       For example:
                       - "chart chart-name" to show chart information.
                       - "values chart-name" to show chart values.
                       - "readme chart-name" to show chart README.

    Returns:
        str: The output of the `helm show` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm show` prefix; it is added automatically.
    """
    return _helm("helm show", args)


def helm_verify(args: str) -> str:
    """
    Run a `helm verify` command and return the output.

    Args:
        args (str): The chart path and options to pass to `helm verify`.
                       For example:
                       - "chart-path" to verify a chart.

    Returns:
        str: The output of the `helm verify` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm verify` prefix; it is added automatically.
    """
    return _helm("helm verify", args)


def helm_env(args: str) -> str:
    """
    Run a `helm env` command and return the output.

    Args:
        args (str): Options to pass to `helm env`.
                       For example:
                       - "" (empty string) to show all environment variables.

    Returns:
        str: The output of the `helm env` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm env` prefix; it is added automatically.
    """
    return _helm("helm env", args)


# ----- Helm RW Commands -----


def helm_install(args: str) -> str:
    """
    Run a `helm install` command and return the output.

    Args:
        args (str): The release name, chart, and options to pass to `helm install`.
                       For example:
                       - "release-name chart-name" to install a chart with the given release name.
                       - "release-name chart-name --set key=value" to install with custom values.
                       - "release-name chart-name -n nginx-system" to install in a specific namespace.

    Returns:
        str: The output of the `helm install` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm install` prefix; it is added automatically.
    """
    return _helm("helm install", args)


def helm_upgrade(args: str) -> str:
    """
    Run a `helm upgrade` command and return the output.

    Args:
        args (str): The release name, chart, and options to pass to `helm upgrade`.
                       For example:
                       - "release-name chart-name" to upgrade a release.
                       - "release-name chart-name --set key=value" to upgrade with custom values.
                       - "release-name chart-name -n nginx-system" to upgrade in a specific namespace.

    Returns:
        str: The output of the `helm upgrade` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm upgrade` prefix; it is added automatically.
    """
    return _helm("helm upgrade", args)


def helm_rollback(args: str) -> str:
    """
    Run a `helm rollback` command and return the output.

    Args:
        args (str): The release name, revision, and options to pass to `helm rollback`.
                       For example:
                       - "release-name 1" to rollback to revision 1.
                       - "release-name 2 -n nginx-system" to rollback in a specific namespace.

    Returns:
        str: The output of the `helm rollback` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm rollback` prefix; it is added automatically.
    """
    return _helm("helm rollback", args)


def helm_uninstall(args: str) -> str:
    """
    Run a `helm uninstall` command and return the output.

    Args:
        args (str): The release name and options to pass to `helm uninstall`.
                       For example:
                       - "release-name" to uninstall a release.
                       - "release-name -n nginx-system" to uninstall from a specific namespace.

    Returns:
        str: The output of the `helm uninstall` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm uninstall` prefix; it is added automatically.
    """
    return _helm("helm uninstall", args)


def helm_test(args: str) -> str:
    """
    Run a `helm test` command and return the output.

    Args:
        args (str): The release name and options to pass to `helm test`.
                       For example:
                       - "release-name" to test a release.
                       - "release-name -n nginx-system" to test a release in a specific namespace.

    Returns:
        str: The output of the `helm test` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm test` prefix; it is added automatically.
    """
    return _helm("helm test", args)


# ----- Helm Admin Commands -----


def helm_repo(args: str) -> str:
    """
    Run a `helm repo` command and return the output.

    Args:
        args (str): The subcommand and options to pass to `helm repo`.
                       For example:
                       - "list" to list chart repositories.
                       - "add repo-name https://charts.example.com/" to add a repository.
                       - "remove repo-name" to remove a repository.
                       - "update" to update information of available charts.

    Returns:
        str: The output of the `helm repo` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm repo` prefix; it is added automatically.
    """
    return _helm("helm repo", args)


def helm_push(args: str) -> str:
    """
    Run a `helm push` command and return the output.

    Args:
        args (str): The chart package and options to pass to `helm push`.
                       For example:
                       - "chart.tgz repo-name" to push a chart to a repository.

    Returns:
        str: The output of the `helm push` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm push` prefix; it is added automatically.
        - This command may require the helm-push plugin to be installed.
    """
    return _helm("helm push", args)


def helm_dependency(args: str) -> str:
    """
    Run a `helm dependency` command and return the output.

    Args:
        args (str): The subcommand and options to pass to `helm dependency`.
                       For example:
                       - "update chart-path" to update chart dependencies.
                       - "list chart-path" to list chart dependencies.

    Returns:
        str: The output of the `helm dependency` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm dependency` prefix; it is added automatically.
    """
    return _helm("helm dependency", args)


def helm_package(args: str) -> str:
    """
    Run a `helm package` command and return the output.

    Args:
        args (str): The chart path and options to pass to `helm package`.
                       For example:
                       - "chart-path" to package a chart directory into a chart archive.
                       - "chart-path --destination ./charts" to specify a destination directory.

    Returns:
        str: The output of the `helm package` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm package` prefix; it is added automatically.
    """
    return _helm("helm package", args)


def helm_registry(args: str) -> str:
    """
    Run a `helm registry` command and return the output.

    Args:
        args (str): The subcommand and options to pass to `helm registry`.
                       For example:
                       - "login registry-url" to login to an OCI registry.
                       - "logout registry-url" to logout from an OCI registry.

    Returns:
        str: The output of the `helm registry` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm registry` prefix; it is added automatically.
    """
    return _helm("helm registry", args)


def helm_pull(args: str) -> str:
    """
    Run a `helm pull` command and return the output.

    Args:
        args (str): The chart and options to pass to `helm pull`.
                       For example:
                       - "repo/chart-name" to download a chart from a repository.
                       - "repo/chart-name --version 1.0.0" to download a specific version.
                       - "repo/chart-name --untar" to unpack the chart after downloading.

    Returns:
        str: The output of the `helm pull` command, or an error message if the command is invalid.

    Notes:
        - The `args` argument should not include the `helm pull` prefix; it is added automatically.
    """
    return _helm("helm pull", args)
