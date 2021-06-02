import './Nethereum.Generators.DuoCode';

function buildConstructor(item: any): Nethereum.Generators.Model.ConstructorABI {
  const constructorItem = new Nethereum.Generators.Model.ConstructorABI();
  constructorItem.set_InputParameters(buildFunctionParameters(item.inputs));
  return constructorItem;
}

function buildFunction(item: any): Nethereum.Generators.Model.FunctionABI {
  const functionItem = new Nethereum.Generators.Model.FunctionABI(item.name, item.constant, false);
  functionItem.set_InputParameters(buildFunctionParameters(item.inputs));
  functionItem.set_OutputParameters(buildFunctionParameters(item.outputs));
  return functionItem;
}

function buildEvent(item: any): Nethereum.Generators.Model.EventABI {
  const eventItem = new Nethereum.Generators.Model.EventABI (item.name);
  eventItem.set_InputParameters(buildEventParameters(item.inputs));
  return eventItem;
}

function buildFunctionParameters(items: any): Nethereum.Generators.Model.ParameterABI[] {
  let parameterOrder = 0;
  const parameters = [];
  for (let i = 0, len = items.length; i < len; i++) {
    parameterOrder = parameterOrder + 1;
    const parameter = new Nethereum.Generators.Model.ParameterABI
      .ctor$1(items[i].type, items[i].name, parameterOrder);
    parameters.push(parameter);
  }
  return parameters;
}

function buildEventParameters(items: any): Nethereum.Generators.Model.ParameterABI[] {
  let parameterOrder = 0;
  const parameters = [];
  for (let i = 0, len = items.length; i < len; i++) {
    parameterOrder = parameterOrder + 1;
    const parameter = new Nethereum.Generators.Model.ParameterABI
      .ctor$1(items[i].type, items[i].name, parameterOrder);
    parameter.set_Indexed(items[i].indexed);
    parameters.push(parameter);
  }
  return parameters;
}

export function buildContract(abiStr: string): Nethereum.Generators.Model.ContractABI {
  const abi = JSON.parse(abiStr);
  const functions = [];
  const events = [];
  let constructor = new Nethereum.Generators.Model.ConstructorABI();

  for (let i = 0, len = abi.length; i < len; i++) {
    if (abi[i].type === 'function') {
      functions.push(buildFunction(abi[i]));
    }

    if (abi[i].type === 'event') {
      events.push(buildEvent(abi[i]));
    }

    if (abi[i].type === 'constructor') {
      constructor = buildConstructor(abi[i]);
    }
  }

  const contract = new Nethereum.Generators.Model.ContractABI();
  contract.set_Constructor(constructor);
  contract.set_Functions(functions);
  contract.set_Events(events);
  return contract;
}
