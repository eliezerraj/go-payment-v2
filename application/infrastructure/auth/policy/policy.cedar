//
//WHO    = Principal (client_id)
//WHAT   = Action (inventory create)
//WHICH  = Resource (order order) when empty means all
//

permit(
    principal,
    action == Action::"get:payment",
    resource
)
when {
    principal.scopes.contains(
        "payment:read"
    )
};

permit(
    principal,
    action == Action::"get:payment",
    resource
)
when {
    principal.roles.contains(
        "order-all-role"
    )
};

permit(
    principal,
    action == Action::"add:payment",
    resource
)
when {
    principal.scopes.contains(
        "payment:create"
    )
};

permit(
    principal,
    action == Action::"add:payment",
    resource
)
when {
    principal.roles.contains(
        "order-all-role"
    )
};

permit(
    principal,
    action == Action::"list:payment",
    resource
)
when {
    principal.scopes.contains(
        "payment:read"
    )
};

permit(
    principal,
    action == Action::"list:payment",
    resource
)
when {
    principal.roles.contains(
        "order-all-role"
    )
};