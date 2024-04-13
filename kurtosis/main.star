el_cl_genesis_data_generator = import_module(
    "github.com/kurtosis-tech/ethereum-package/src/prelaunch_data_generator/el_cl_genesis/el_cl_genesis_generator.star",
)

execution = import_module("./src/nodes/execution/execution.star")
execution_types = import_module("./src/nodes/execution/types.star")
beacond = import_module("./src/nodes/consensus/beacond/launcher.star")
networks = import_module("./src/networks/networks.star")
port_spec_lib = import_module("./src/lib/port_spec.star")
nodes = import_module("./src/nodes/nodes.star")
nginx = import_module("./src/services/nginx/nginx.star")
constants = import_module("./src/constants.star")
goomy_blob = import_module("./src/services/goomy/launcher.star")
init = import_module("./src/lib/init.star")

def run(plan, validators, full_nodes = [], rpc_endpoints = [], additional_services = []):
    """
    Initiates the execution plan with the specified number of validators and arguments.

    Args:
    plan: The execution plan to be run.
    args: Additional arguments to configure the plan. Defaults to an empty dictionary.
    """

    validators = nodes.parse_nodes_from_dict(validators)
    full_nodes = nodes.parse_nodes_from_dict(full_nodes)
    num_validators = len(validators)

    # 1. Initialize EVM genesis data
    evm_genesis_data = networks.get_genesis_data(plan)

    node_modules = {}
    for node in validators:
        if node.el_type not in node_modules.keys():
            node_path = "./src/nodes/execution/{}/config.star".format(node.el_type)
            node_module = import_module(node_path)
            node_modules[node.el_type] = node_module

    # 2. Upload files
    jwt_file, kzg_trusted_setup = execution.upload_global_files(plan, node_modules)

    # 3. Perform genesis ceremony
    node_peering_info = beacond.perform_genesis_ceremony(plan, validators, jwt_file)

    el_enode_addrs = []

    # 4. Start network validators
    for n, validator in enumerate(validators):
        el_client = execution.create_node(plan, node_modules, validator, "validator", n, el_enode_addrs)
        el_enode_addrs.append(el_client["enode_addr"])

        # 4b. Launch CL
        beacond.create_node(plan, validator.cl_image, node_peering_info[:n], el_client["name"], jwt_file, kzg_trusted_setup, n)

    # 5. Start full nodes (rpcs)
    full_node_configs = {}
    for n, full in enumerate(full_nodes):
        el_client = execution.create_node(plan, node_modules, full, "full", n, el_enode_addrs)
        el_enode_addrs.append(el_client["enode_addr"])

        # 4b. Launch CL
        cl_service_name = "cl-full-beaconkit-{}".format(n)
        full_node_config = beacond.create_full_node_config(plan, full.cl_image, node_peering_info, el_client["name"], jwt_file, kzg_trusted_setup, n)
        full_node_configs[cl_service_name] = full_node_config

    if full_node_configs != {}:
        plan.add_services(
            configs = full_node_configs,
        )

    # for cl_service_name, full_node_config in full_node_configs.items():
    #     init.init_beacond(plan, full_node_config.env_vars["BEACOND_CHAIN_ID"], full_node_config.env_vars["BEACOND_MONIKER"],full_node_config.env_vars["BEACOND_HOME"], False, cl_service_name)
    # 6. Start RPCs
    for n, rpc in enumerate(rpc_endpoints):
        nginx.get_config(plan, rpc["services"])

    # 7. Start additional services
    for s in additional_services:
        if s == "goomy_blob":
            plan.print("Launching Goomy the Blob Spammer")
            goomy_blob_args = {"goomy_blob_args": []}
            goomy_blob.launch_goomy_blob(
                plan,
                constants.PRE_FUNDED_ACCOUNTS[0],
                plan.get_service("nginx").ports["http"].url,
                goomy_blob_args,
            )
            plan.print("Successfully launched goomy the blob spammer")

    plan.print("Successfully launched development network")
